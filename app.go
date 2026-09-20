package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	goRuntime "runtime"
	"strconv"
	"strings"

	"github.com/OmarNaru1110/byteless/internal/command"
	"github.com/OmarNaru1110/byteless/internal/config"
	"github.com/OmarNaru1110/byteless/internal/domain"
	"github.com/OmarNaru1110/byteless/internal/util"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	inputVideo      *domain.Video
	compressedVideo *domain.OutputVideo
	cancel          context.CancelFunc
	tools           config.Config
}

const (
	minVideoBitrateKbps    = 100
	outputAudioBitrateKbps = 128
)

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		tools: config.Default(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetDefaultOutputDir returns the default folder for compressed videos
func (a *App) GetDefaultOutputDir() string {
	if a.compressedVideo == nil {
		a.compressedVideo = &domain.OutputVideo{}
	}
	a.compressedVideo.OutputPath = a.defaultOutputDir()
	log.Printf("App: GetDefaultOutputDir -> %q", a.compressedVideo.OutputPath)
	return a.compressedVideo.OutputPath
}

func (a *App) defaultOutputDir() string {
	if a.inputVideo != nil {
		if fi, err := os.Stat(a.inputVideo.Path); err == nil && fi.IsDir() {
			log.Printf("App: defaultOutputDir using input video dir %q", a.inputVideo.Path)
			return a.inputVideo.Path
		}
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		log.Printf("App: defaultOutputDir: user home dir unavailable (err=%v), using '.'", err)
		return "."
	}

	for _, name := range []string{"Videos", "Downloads"} {
		dir := filepath.Join(home, name)
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			log.Printf("App: defaultOutputDir using %q", dir)
			return dir
		}
	}

	log.Printf("App: defaultOutputDir falling back to home dir %q", home)
	return home
}

// PickFolder opens a native directory picker dialog seeded with defaultDir
func (a *App) PickFolder(defaultDir string) (string, error) {
	log.Printf("App: PickFolder opening dialog seeded with %q", defaultDir)
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Choose destination folder",
		DefaultDirectory:     defaultDir,
		CanCreateDirectories: true,
	})
	if err != nil {
		log.Printf("App: PickFolder failed: %v", err)
		return "", err
	}
	log.Printf("App: PickFolder chose %q", path)
	return path, nil
}

// SelectVideoFile opens a native video file picker and returns the chosen path.
// The caller is expected to load the video afterwards so the UI can show
// loading feedback while the file is probed.
func (a *App) SelectVideoFile() (string, error) {
	log.Println("App: SelectVideoFile opening file dialog")
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose a video to compress",
		Filters: []runtime.FileFilter{
			{DisplayName: "Video files", Pattern: "*.mp4;*.mov;*.mkv;*.webm"},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
	if err != nil {
		log.Printf("App: SelectVideoFile failed: %v", err)
		return "", err
	}
	if path == "" {
		log.Println("App: SelectVideoFile cancelled by user")
	}
	log.Printf("App: SelectVideoFile picked %q", path)
	return path, nil
}

// LoadVideo probes the video at path, stores it as the input, and returns its details
func (a *App) LoadVideo(path string) (*domain.Video, error) {
	log.Printf("App: LoadVideo called with %q", path)
	video, err := a.probeVideo(path)
	if err != nil {
		log.Printf("App: LoadVideo failed for %q: %v", path, err)
		return nil, err
	}

	a.inputVideo = video
	if a.compressedVideo == nil {
		a.compressedVideo = &domain.OutputVideo{}
	}
	a.compressedVideo.OutputPath = a.defaultOutputDir()

	log.Printf("App: LoadVideo loaded video name=%q size=%d duration=%d audioBitrate=%d", video.Name, video.Size, video.Duration, video.AudioBitrate)
	return video, nil
}

func (a *App) probeVideo(path string) (*domain.Video, error) {
	log.Printf("App: probeVideo probing %q", path)
	fi, err := os.Stat(path)
	if err != nil {
		log.Printf("App: probeVideo stat failed for %q: %v", path, err)
		return nil, err
	}
	if fi.IsDir() {
		log.Printf("App: probeVideo: %q is a directory", path)
		return nil, fmt.Errorf("path is a directory, expected a video file")
	}

	cmd, err := command.NewGetVideoDetailsCommand(a.tools.FFprobePath, path)
	if err != nil {
		log.Printf("App: probeVideo failed to build details command: %v", err)
		return nil, err
	}

	details, err := cmd.Execute()
	if err != nil {
		log.Printf("App: probeVideo ffprobe failed: %v", err)
		return nil, err
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(details.Format.Duration), 64)
	if err != nil {
		log.Printf("App: probeVideo failed to parse duration %q, defaulting to 0", details.Format.Duration)
		duration = 0
	}

	var audioBitrate int
	if len(details.Streams) > 0 {
		if bitrate, err := strconv.Atoi(strings.TrimSpace(details.Streams[0].BitRate)); err == nil {
			audioBitrate = bitrate / 1000 // bps -> kbps
		}
	}

	log.Printf("App: probeVideo done for %q: duration=%.2fs audioBitrate=%dkbps", path, duration, audioBitrate)
	return &domain.Video{
		Path:         filepath.Dir(path),
		Name:         filepath.Base(path),
		Size:         int(fi.Size()),
		Duration:     int(duration),
		AudioBitrate: audioBitrate,
	}, nil
}

func (a *App) ShowInFolder(path string) error {
	log.Printf("App: ShowInFolder called with %q", path)

	var cmd *exec.Cmd
	switch goRuntime.GOOS {
	case "windows":
		info, err := os.Stat(path)
		if err != nil {
			log.Printf("App: ShowInFolder stat failed for %q: %v", path, err)
			return err
		}
		if info.IsDir() {
			cmd = exec.Command("explorer", path)
		} else {
			cmd = exec.Command("explorer", "/select,", path)
		}
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	case "linux":
		cmd = exec.Command("xdg-open", filepath.Dir(path))
	default:
		log.Printf("App: ShowInFolder unsupported OS %q", goRuntime.GOOS)
		return fmt.Errorf("unsupported OS: %s", goRuntime.GOOS)
	}

	return cmd.Run()
}

func (a *App) OpenFile(path string) error {
	log.Printf("App: OpenFile called with %q", path)

	var cmd *exec.Cmd
	switch goRuntime.GOOS {
	case "windows":
		info, err := os.Stat(path)
		if err != nil {
			log.Printf("App: OpenFile stat failed for %q: %v", path, err)
			return err
		}
		if info.IsDir() {
			cmd = exec.Command("explorer", path)
		} else {
			cmd = exec.Command("cmd", "/C", "start", "", path)
		}
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		log.Printf("App: OpenFile unsupported OS %q", goRuntime.GOOS)
		return fmt.Errorf("unsupported OS: %s", goRuntime.GOOS)
	}

	util.HideWindow(cmd)

	return cmd.Run()
}

// GetEncoders returns the list of available video encoders
func (a *App) GetEncoders() []domain.EncoderInfo {
	log.Println("App: GetEncoders called")
	return domain.GetEncoders()
}

// CancelCompression cancels an in-progress compression
func (a *App) CancelCompression() {
	log.Println("App: CancelCompression called")
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
}

func (a *App) CompressVideo(targetSizeMB float64, outputPath string, encoder domain.VideoEncoder) (string, error) {
	log.Printf("App: CompressVideo called with targetSize=%.2fMB, outputPath=%q, encoder=%q", targetSizeMB, outputPath, encoder)

	if a.inputVideo == nil {
		log.Printf("App: CompressVideo failed: no input video loaded")
		return "", fmt.Errorf("no input video loaded")
	}

	if util.ConvertMBToBytes(targetSizeMB) >= a.inputVideo.Size {
		log.Printf("App: CompressVideo failed: target size %.2fMB is not smaller than input video size %d bytes", targetSizeMB, a.inputVideo.Size)
		return "", fmt.Errorf("target size must be smaller than input video size")
	}

	if outputPath == "" {
		log.Printf("App: CompressVideo failed: output path is empty")
		return "", fmt.Errorf("output path cannot be empty")
	}

	if _, err := os.Stat(outputPath); err != nil {
		log.Printf("App: CompressVideo failed: output path %q does not exist", outputPath)
		return "", fmt.Errorf("output path does not exist")
	}

	a.compressedVideo.TargetSize = targetSizeMB
	a.compressedVideo.OutputPath = outputPath
	a.compressedVideo.Encoder = encoder
	a.compressedVideo.OutputName = util.GenerateName("mp4")

	minPossibleSizeMB, _ := a.GetMinPossibleSize()
	targetSizeAfterMarginMB := util.ApplyMargin(targetSizeMB, 10)
	a.compressedVideo.TargetSizeAfterMargin = max(targetSizeAfterMarginMB, minPossibleSizeMB)

	totalBitrateKbps := (a.compressedVideo.TargetSizeAfterMargin * 8192) / float64(a.inputVideo.Duration)
	targetVideoBitrateKbps := int(totalBitrateKbps - outputAudioBitrateKbps)

	if targetVideoBitrateKbps < minVideoBitrateKbps {
		log.Printf("App: CompressVideo failed: calculated target bitrate %d kbps is not positive", targetVideoBitrateKbps)
		return "", fmt.Errorf("target size is too small")
	}

	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel

	pass1Cmd := command.NewTwoPassEncodePass1Command(a.tools.FFmpegPath, a.inputVideo.FullPath(), targetVideoBitrateKbps, encoder)
	if err := pass1Cmd.Execute(ctx, a.inputVideo.Duration); err != nil {
		log.Printf("App: CompressVideo pass 1 failed: %v", err)
		return "", err
	}

	outputFilePath := a.compressedVideo.FullPath()
	pass2Cmd := command.NewTwoPassEncodePass2Command(a.tools.FFmpegPath, a.inputVideo.FullPath(), targetVideoBitrateKbps, outputFilePath, encoder)
	if err := pass2Cmd.Execute(ctx, a.inputVideo.Duration); err != nil {
		log.Printf("App: CompressVideo pass 2 failed: %v", err)
		return "", err
	}

	log.Printf("App: CompressVideo completed successfully -> %q", outputFilePath)
	return outputFilePath, nil
}

func (a *App) GetMinPossibleSize() (float64, error) {
	if a.inputVideo == nil {
		return 0, fmt.Errorf("no input video loaded")
	}
	minPossibleSizeMB := float64((minVideoBitrateKbps+outputAudioBitrateKbps)*a.inputVideo.Duration) / 8192
	return minPossibleSizeMB, nil
}
