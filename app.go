package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/OmarNaru1110/byteless/internal/command"
	"github.com/OmarNaru1110/byteless/internal/domain"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	inputVideo      *domain.Video
	compressedVideo *domain.OutputVideo
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetDefaultOutputDir returns the default folder for compressed videos
func (a *App) GetDefaultOutputDir() string {
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

// SelectVideoFile opens a native video file picker and stores the chosen input
func (a *App) SelectVideoFile() (*domain.Video, error) {
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
		return nil, err
	}
	if path == "" {
		log.Println("App: SelectVideoFile cancelled by user")
		return nil, nil
	}
	log.Printf("App: SelectVideoFile picked %q", path)
	return a.LoadVideo(path)
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

	cmd, err := command.NewGetVideoDetailsCommand(path)
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
