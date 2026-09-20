package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

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
	original := EmitEvent
	var mu sync.Mutex
	var events []emittedEvent
	EmitEvent = func(_ context.Context, eventName string, data ...interface{}) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, emittedEvent{name: eventName, data: data})
	}
	t.Cleanup(func() { EmitEvent = original })
	return func() []emittedEvent {
		mu.Lock()
		defer mu.Unlock()
		return append([]emittedEvent(nil), events...)
	}
}

func writeTempVideo(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "input.mp4")
	if err := os.WriteFile(p, []byte("fake video content"), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestNewGetVideoDetailsCommandRejectsMissingFile(t *testing.T) {
	if _, err := NewGetVideoDetailsCommand("ffprobe", filepath.Join(t.TempDir(), "missing.mp4")); err == nil {
		t.Fatal("expected error for missing input file")
	}
}

func TestNewGetVideoDetailsCommandAcceptsExistingFile(t *testing.T) {
	file := writeTempVideo(t)
	if _, err := NewGetVideoDetailsCommand("ffprobe", file); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGetVideoDetailsCommandBuildArgs(t *testing.T) {
	file := writeTempVideo(t)
	c, err := NewGetVideoDetailsCommand("ffprobe", file)
	if err != nil {
		t.Fatal(err)
	}
	got := c.builder.Build()
	want := []string{
		"-v", "error",
		"-show_entries", "stream=bit_rate",
		"-show_entries", "format=duration",
		"-show_entries", "format=size",
		"-of", "json",
		file,
		"-select_streams", "a:0",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Build() = %v, want %v", got, want)
	}
}

func TestGetVideoDetailsCommandExecute(t *testing.T) {
	file := writeTempVideo(t)
	c, err := NewGetVideoDetailsCommand(testhelper.FakeMediaBinary(), file)
	if err != nil {
		t.Fatal(err)
	}
	details, err := c.Execute()
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}
	if details.Format.Duration != "100.000000" {
		t.Fatalf("duration = %q, want 100.000000", details.Format.Duration)
	}
	if details.Format.Size != "104857600" {
		t.Fatalf("size = %q, want 104857600", details.Format.Size)
	}
	if len(details.Streams) != 1 {
		t.Fatalf("got %d streams, want 1", len(details.Streams))
	}
	if details.Streams[0].BitRate != "128000" {
		t.Fatalf("bit_rate = %q, want 128000", details.Streams[0].BitRate)
	}
}

func TestGetVideoDetailsCommandExecuteFailure(t *testing.T) {
	t.Setenv(testhelper.MediaHelperFailEnvKey, "1")
	file := writeTempVideo(t)
	c, err := NewGetVideoDetailsCommand(testhelper.FakeMediaBinary(), file)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Execute(); err == nil {
		t.Fatal("expected error from failing fake binary")
	}
}

func TestGetVideoDetailsCommandNilBuilder(t *testing.T) {
	c := &GetVideoDetailsCommand{}
	if _, err := c.Execute(); err == nil {
		t.Fatal("expected error for nil builder")
	}
}

func TestTwoPassEncodePass1CommandBuildArgs(t *testing.T) {
	c := NewTwoPassEncodePass1Command("ffmpeg", "in.mp4", 1500, domain.H264)
	got := c.builder.Build()
	want := []string{
		"-i", "in.mp4",
		"-c:v", "libx264",
		"-b:v", "1500k",
		"-pass", "1",
		"-an",
		"-f", "null",
		"-progress", "pipe:1",
		"NUL",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Build() = %v, want %v", got, want)
	}
}

func TestTwoPassEncodePass1CommandExecuteEmitsProgress(t *testing.T) {
	events := captureEmittedEvents(t)
	c := NewTwoPassEncodePass1Command(testhelper.FakeMediaBinary(), "in.mp4", 1500, domain.H264)
	if err := c.Execute(context.Background(), 10); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}
	evs := events()
	if len(evs) != 1 {
		t.Fatalf("got %d emitted events, want 1: %+v", len(evs), evs)
	}
	if evs[0].name != "pass1Progress" {
		t.Fatalf("event name = %q, want pass1Progress", evs[0].name)
	}
	if len(evs[0].data) != 1 || fmt.Sprint(evs[0].data[0]) != "5" {
		t.Fatalf("event data = %v, want [5]", evs[0].data)
	}
}

func TestTwoPassEncodePass1CommandExecuteFailure(t *testing.T) {
	t.Setenv(testhelper.MediaHelperFailEnvKey, "1")
	c := NewTwoPassEncodePass1Command(testhelper.FakeMediaBinary(), "in.mp4", 1500, domain.H264)
	if err := c.Execute(context.Background(), 10); err == nil {
		t.Fatal("expected error from failing fake binary")
	}
}

func TestTwoPassEncodePass1CommandNilBuilder(t *testing.T) {
	c := &TwoPassEncodePass1Command{}
	if err := c.Execute(context.Background(), 10); err == nil {
		t.Fatal("expected error for nil builder")
	}
}

func TestTwoPassEncodePass2CommandBuildArgs(t *testing.T) {
	c := NewTwoPassEncodePass2Command("ffmpeg", "in.mp4", 1500, "out.mp4", domain.H265)
	got := c.builder.Build()
	want := []string{
		"-i", "in.mp4",
		"-c:v", "libx265",
		"-b:v", "1500k",
		"-pass", "2",
		"-c:a", "aac",
		"-b:a", "128k",
		"-progress", "pipe:1",
		"out.mp4",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Build() = %v, want %v", got, want)
	}
}

func TestTwoPassEncodePass2CommandExecuteEmitsProgress(t *testing.T) {
	events := captureEmittedEvents(t)
	c := NewTwoPassEncodePass2Command(testhelper.FakeMediaBinary(), "in.mp4", 1500, "out.mp4", domain.H264)
	if err := c.Execute(context.Background(), 10); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}
	evs := events()
	if len(evs) != 1 {
		t.Fatalf("got %d emitted events, want 1: %+v", len(evs), evs)
	}
	if evs[0].name != "pass2Progress" {
		t.Fatalf("event name = %q, want pass2Progress", evs[0].name)
	}
	if len(evs[0].data) != 1 || fmt.Sprint(evs[0].data[0]) != "5" {
		t.Fatalf("event data = %v, want [5]", evs[0].data)
	}
}

func TestTwoPassEncodePass2CommandExecuteFailure(t *testing.T) {
	t.Setenv(testhelper.MediaHelperFailEnvKey, "1")
	c := NewTwoPassEncodePass2Command(testhelper.FakeMediaBinary(), "in.mp4", 1500, "out.mp4", domain.H264)
	if err := c.Execute(context.Background(), 10); err == nil {
		t.Fatal("expected error from failing fake binary")
	}
}

func TestTwoPassEncodePass2CommandNilBuilder(t *testing.T) {
	c := &TwoPassEncodePass2Command{}
	if err := c.Execute(context.Background(), 10); err == nil {
		t.Fatal("expected error for nil builder")
	}
}