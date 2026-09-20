package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	goRuntime "runtime"
	"strings"
	"sync"
	"testing"

	"github.com/OmarNaru1110/byteless/internal/command"
	"github.com/OmarNaru1110/byteless/internal/config"
	"github.com/OmarNaru1110/byteless/internal/domain"
	"github.com/OmarNaru1110/byteless/internal/testhelper"
)

func TestMain(m *testing.M) {
	if testhelper.IsMediaHelper() {
		testhelper.HelperMain()
	}
	testhelper.SetMediaHelperEnv()
	os.Exit(m.Run())
}

type emittedEvent struct {
	name string
	data []interface{}
}

func captureEmittedEvents(t *testing.T) func() []emittedEvent {
	t.Helper()
	original := command.EmitEvent
	var mu sync.Mutex
	var events []emittedEvent
	command.EmitEvent = func(_ context.Context, eventName string, data ...interface{}) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, emittedEvent{name: eventName, data: data})
	}
	t.Cleanup(func() { command.EmitEvent = original })
	return func() []emittedEvent {
		mu.Lock()
		defer mu.Unlock()
		return append([]emittedEvent(nil), events...)
	}
}

func setHomeDir(t *testing.T, dir string) {
	t.Helper()
	if goRuntime.GOOS == "windows" {
		t.Setenv("USERPROFILE", dir)
	} else {
		t.Setenv("HOME", dir)
	}
}

func setEmptyHomeEnv(t *testing.T) {
	t.Helper()
	if goRuntime.GOOS == "windows" {
		t.Setenv("USERPROFILE", "")
		t.Setenv("HOMEDRIVE", "")
		t.Setenv("HOMEPATH", "")
		t.Setenv("HOME", "")
	} else {
		t.Setenv("HOME", "")
	}
}

func writeVideoFile(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, content, 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestNewApp(t *testing.T) {
	a := NewApp()
	if a.tools == (config.Config{}) {
		t.Fatal("expected tools to be configured")
	}
	if a.inputVideo != nil {
		t.Fatal("expected inputVideo to be nil for a fresh app")
	}
	if a.compressedVideo != nil {
		t.Fatal("expected compressedVideo to be nil for a fresh app")
	}
	if a.ctx != nil {
		t.Fatal("expected ctx to be set only after startup")
	}
}

func TestStartup(t *testing.T) {
	a := NewApp()
	ctx := context.Background()
	a.startup(ctx)
	if a.ctx != ctx {
		t.Fatalf("startup did not store the context")
	}
}

func TestGetMinPossibleSizeRequiresInput(t *testing.T) {
	a := &App{}
	if _, err := a.GetMinPossibleSize(); err == nil {
		t.Fatal("expected error without an input video")
	}
}

func TestGetMinPossibleSize(t *testing.T) {
	a := &App{inputVideo: &domain.Video{Duration: 100}}
	got, err := a.GetMinPossibleSize()
	if err != nil {
		t.Fatal(err)
	}
	want := float64((minVideoBitrateKbps+outputAudioBitrateKbps)*100) / 8192
	if got != want {
		t.Fatalf("GetMinPossibleSize() = %v, want %v", got, want)
	}
}

func TestDefaultOutputDirUsesInputDir(t *testing.T) {
	dir := t.TempDir()
	a := &App{inputVideo: &domain.Video{Path: dir, Name: "clip.mp4"}}
	if got := a.defaultOutputDir(); got != dir {
		t.Fatalf("defaultOutputDir() = %q, want %q", got, dir)
	}
}

func TestDefaultOutputDirFallsBackToHomeWhenInputDirMissing(t *testing.T) {
	home := t.TempDir()
	setHomeDir(t, home)
	a := &App{inputVideo: &domain.Video{Path: filepath.Join(t.TempDir(), "does-not-exist"), Name: "clip.mp4"}}
	if got := a.defaultOutputDir(); got != home {
		t.Fatalf("defaultOutputDir() = %q, want home %q", got, home)
	}
}

func TestDefaultOutputDirPrefersVideos(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "Videos"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "Downloads"), 0755); err != nil {
		t.Fatal(err)
	}
	setHomeDir(t, home)
	a := &App{}
	if got := a.defaultOutputDir(); got != filepath.Join(home, "Videos") {
		t.Fatalf("defaultOutputDir() = %q, want %q", got, filepath.Join(home, "Videos"))
	}
}

func TestDefaultOutputDirPrefersDownloads(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "Downloads"), 0755); err != nil {
		t.Fatal(err)
	}
	setHomeDir(t, home)
	a := &App{}
	if got := a.defaultOutputDir(); got != filepath.Join(home, "Downloads") {
		t.Fatalf("defaultOutputDir() = %q, want %q", got, filepath.Join(home, "Downloads"))
	}
}

func TestDefaultOutputDirFallsBackToHome(t *testing.T) {
	home := t.TempDir()
	setHomeDir(t, home)
	a := &App{}
	if got := a.defaultOutputDir(); got != home {
		t.Fatalf("defaultOutputDir() = %q, want home %q", got, home)
	}
}

func TestDefaultOutputDirUsesDotWhenHomeUnavailable(t *testing.T) {
	setEmptyHomeEnv(t)
	a := &App{inputVideo: &domain.Video{Path: filepath.Join(t.TempDir(), "does-not-exist")}}
	if got := a.defaultOutputDir(); got != "." {
		t.Fatalf("defaultOutputDir() = %q, want .", got)
	}
}

func TestGetDefaultOutputDir(t *testing.T) {
	dir := t.TempDir()
	a := &App{
		inputVideo:      &domain.Video{Path: dir, Name: "clip.mp4"},
		compressedVideo: &domain.OutputVideo{},
	}
	if got := a.GetDefaultOutputDir(); got != dir {
		t.Fatalf("GetDefaultOutputDir() = %q, want %q", got, dir)
	}
	if a.compressedVideo.OutputPath != dir {
		t.Fatalf("OutputPath = %q, want %q", a.compressedVideo.OutputPath, dir)
	}
}

func TestLoadVideoRejectsMissingFile(t *testing.T) {
	a := NewApp()
	if _, err := a.LoadVideo(filepath.Join(t.TempDir(), "missing.mp4")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadVideoRejectsDirectory(t *testing.T) {
	a := NewApp()
	if _, err := a.LoadVideo(t.TempDir()); err == nil {
		t.Fatal("expected error for directory input")
	}
}

func TestLoadVideoSuccess(t *testing.T) {
	dir := t.TempDir()
	path := writeVideoFile(t, dir, "movie.mp4", []byte(strings.Repeat("x", 512)))

	a := NewApp()
	a.tools.FFprobePath = testhelper.FakeMediaBinary()

	v, err := a.LoadVideo(path)
	if err != nil {
		t.Fatalf("LoadVideo() failed: %v", err)
	}
	if v.Name != "movie.mp4" {
		t.Fatalf("Name = %q, want movie.mp4", v.Name)
	}
	if v.Path != dir {
		t.Fatalf("Path = %q, want %q", v.Path, dir)
	}
	if v.Size != 512 {
		t.Fatalf("Size = %d, want 512", v.Size)
	}
	if v.Duration != 100 {
		t.Fatalf("Duration = %d, want 100", v.Duration)
	}
	if v.AudioBitrate != 128 {
		t.Fatalf("AudioBitrate = %d, want 128", v.AudioBitrate)
	}
	if a.inputVideo != v {
		t.Fatal("LoadVideo did not store the input video")
	}
	if a.compressedVideo == nil {
		t.Fatal("LoadVideo did not initialize compressedVideo")
	}
	if a.compressedVideo.OutputPath != dir {
		t.Fatalf("default output path = %q, want %q", a.compressedVideo.OutputPath, dir)
	}
}

func TestCompressVideoRequiresInput(t *testing.T) {
	a := &App{}
	if _, err := a.CompressVideo(10, t.TempDir(), domain.H264); err == nil {
		t.Fatal("expected error without an input video")
	}
}

func TestCompressVideoRejectsTargetSizeNotSmallerThanInput(t *testing.T) {
	a := &App{inputVideo: &domain.Video{Size: 10 * 1024 * 1024}}
	if _, err := a.CompressVideo(10, t.TempDir(), domain.H264); err == nil {
		t.Fatal("expected error for equal target size")
	}
	if _, err := a.CompressVideo(11, t.TempDir(), domain.H264); err == nil {
		t.Fatal("expected error for larger target size")
	}
}

func TestCompressVideoRequiresOutputPath(t *testing.T) {
	a := &App{inputVideo: &domain.Video{Size: 10 * 1024 * 1024}}
	if _, err := a.CompressVideo(5, "", domain.H264); err == nil {
		t.Fatal("expected error for empty output path")
	}
}

func TestCompressVideoRequiresExistingOutputDir(t *testing.T) {
	a := &App{inputVideo: &domain.Video{Size: 10 * 1024 * 1024}}
	if _, err := a.CompressVideo(5, filepath.Join(t.TempDir(), "missing"), domain.H264); err == nil {
		t.Fatal("expected error for missing output dir")
	}
}

func TestCompressVideoSuccess(t *testing.T) {
	events := captureEmittedEvents(t)

	dir := t.TempDir()
	writeVideoFile(t, dir, "input.mp4", []byte("fake video content"))

	outputDir := t.TempDir()
	a := NewApp()
	a.ctx = context.Background()
	a.tools.FFmpegPath = testhelper.FakeMediaBinary()
	a.inputVideo = &domain.Video{
		Path:         dir,
		Name:         "input.mp4",
		Size:         100 * 1024 * 1024,
		Duration:     100,
		AudioBitrate: 128,
	}
	a.compressedVideo = &domain.OutputVideo{}

	outPath, err := a.CompressVideo(50, outputDir, domain.H264)
	if err != nil {
		t.Fatalf("CompressVideo() failed: %v", err)
	}
	if outPath == "" {
		t.Fatal("CompressVideo() returned an empty output path")
	}
	if outPath != a.compressedVideo.FullPath() {
		t.Fatalf("output path = %q, want %q", outPath, a.compressedVideo.FullPath())
	}
	if filepath.Dir(outPath) != outputDir {
		t.Fatalf("output dir = %q, want %q", filepath.Dir(outPath), outputDir)
	}
	if filepath.Ext(outPath) != ".mp4" {
		t.Fatalf("output extension = %q, want .mp4", filepath.Ext(outPath))
	}
	if !regexp.MustCompile(`^byteless_\d+\.mp4$`).MatchString(filepath.Base(outPath)) {
		t.Fatalf("output name = %q, want byteless_<timestamp>.mp4", filepath.Base(outPath))
	}
	if a.compressedVideo.TargetSize != 50 {
		t.Fatalf("TargetSize = %v, want 50", a.compressedVideo.TargetSize)
	}
	if a.compressedVideo.TargetSizeAfterMargin != 45 {
		t.Fatalf("TargetSizeAfterMargin = %v, want 45", a.compressedVideo.TargetSizeAfterMargin)
	}
	if a.compressedVideo.Encoder != domain.H264 {
		t.Fatalf("Encoder = %q, want %q", a.compressedVideo.Encoder, domain.H264)
	}

	emitted := events()
	names := make(map[string]bool)
	for _, e := range emitted {
		names[e.name] = true
	}
	for _, want := range []string{"pass1Progress", "pass2Progress"} {
		if !names[want] {
			t.Fatalf("expected %s event, got %+v", want, emitted)
		}
	}
}

func TestCancelCompression(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	a := &App{cancel: cancel}
	a.CancelCompression()
	if a.cancel != nil {
		t.Fatal("expected cancel function to be cleared")
	}
	if ctx.Err() != context.Canceled {
		t.Fatal("expected the cancel function to be invoked")
	}
	a.CancelCompression()
}

func TestCancelCompressionNoopWhenNoCompression(t *testing.T) {
	a := &App{}
	a.CancelCompression()
}

func TestGetEncoders(t *testing.T) {
	a := &App{}
	encoders := a.GetEncoders()
	if len(encoders) != 2 {
		t.Fatalf("GetEncoders() returned %d encoders, want 2", len(encoders))
	}
	if encoders[0].ID != domain.H264 || encoders[1].ID != domain.H265 {
		t.Fatalf("GetEncoders() = %+v, want [H264 H265]", encoders)
	}
}

func TestShowInFolderMissingPath(t *testing.T) {
	if goRuntime.GOOS != "windows" {
		t.Skip("windows-only stat-checked behavior")
	}
	a := &App{}
	if err := a.ShowInFolder(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestOpenFileMissingPath(t *testing.T) {
	if goRuntime.GOOS != "windows" {
		t.Skip("windows-only stat-checked behavior")
	}
	a := &App{}
	if err := a.OpenFile(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error for missing path")
	}
}