package builder

import (
	"reflect"
	"testing"
)

func TestFfmpegBuilderGetFfmpegPath(t *testing.T) {
	if got := NewFfmpegBuilder("C:\\tools\\ffmpeg.exe").GetFfmpegPath(); got != "C:\\tools\\ffmpeg.exe" {
		t.Fatalf("GetFfmpegPath() = %q", got)
	}
}

func TestFfmpegBuilderMethodsChain(t *testing.T) {
	b := NewFfmpegBuilder("ffmpeg")
	if b.SetInputFilePath("in") != b {
		t.Fatal("SetInputFilePath must return the same builder")
	}
	if b.SetOutputFilePath("out") != b {
		t.Fatal("SetOutputFilePath must return the same builder")
	}
	if b.SetVideoCodec("libx264") != b {
		t.Fatal("SetVideoCodec must return the same builder")
	}
	if b.SetVideoBitrate("1500k") != b {
		t.Fatal("SetVideoBitrate must return the same builder")
	}
	if b.SetPass(1) != b {
		t.Fatal("SetPass must return the same builder")
	}
	if b.SetAudioCodec("aac") != b {
		t.Fatal("SetAudioCodec must return the same builder")
	}
	if b.SetAudioBitrate("128k") != b {
		t.Fatal("SetAudioBitrate must return the same builder")
	}
	if b.DisableAudio() != b {
		t.Fatal("DisableAudio must return the same builder")
	}
	if b.SetFormat("null") != b {
		t.Fatal("SetFormat must return the same builder")
	}
	if b.SetProgressPipe("pipe:1") != b {
		t.Fatal("SetProgressPipe must return the same builder")
	}
}

func TestFfmpegBuilderBuildEmptyUsesNUL(t *testing.T) {
	got := NewFfmpegBuilder("ffmpeg").Build()
	if want := []string{"NUL"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Build() = %v, want %v", got, want)
	}
}

func TestFfmpegBuilderBuildWithOutputPath(t *testing.T) {
	got := NewFfmpegBuilder("ffmpeg").
		SetInputFilePath("in.mp4").
		SetOutputFilePath("out.mp4").
		Build()
	want := []string{"-i", "in.mp4", "out.mp4"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Build() = %v, want %v", got, want)
	}
}

func TestFfmpegBuilderBuildAllOptions(t *testing.T) {
	got := NewFfmpegBuilder("C:\\tools\\ffmpeg.exe").
		SetInputFilePath("in.mp4").
		SetVideoCodec("libx264").
		SetVideoBitrate("1500k").
		SetPass(1).
		SetAudioCodec("aac").
		SetAudioBitrate("128k").
		DisableAudio().
		SetFormat("null").
		SetProgressPipe("pipe:1").
		SetOutputFilePath("out.mp4").
		Build()
	want := []string{
		"-i", "in.mp4",
		"-c:v", "libx264",
		"-b:v", "1500k",
		"-pass", "1",
		"-c:a", "aac",
		"-b:a", "128k",
		"-an",
		"-f", "null",
		"-progress", "pipe:1",
		"out.mp4",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Build() = %v, want %v", got, want)
	}
}

func TestFfmpegBuilderBuildWithoutOutputPath(t *testing.T) {
	got := NewFfmpegBuilder("ffmpeg").
		SetInputFilePath("in.mp4").
		SetVideoCodec("libx265").
		SetVideoBitrate("1000k").
		SetPass(2).
		Build()
	want := []string{"-i", "in.mp4", "-c:v", "libx265", "-b:v", "1000k", "-pass", "2", "NUL"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Build() = %v, want %v", got, want)
	}
}

func TestFfprobeBuilderGetFfprobePath(t *testing.T) {
	if got := NewFfprobeBuilder("C:\\tools\\ffprobe.exe").GetFfprobePath(); got != "C:\\tools\\ffprobe.exe" {
		t.Fatalf("GetFfprobePath() = %q", got)
	}
}

func TestFfprobeBuilderMethodsChain(t *testing.T) {
	b := NewFfprobeBuilder("ffprobe")
	if b.SetLogLevel("error") != b {
		t.Fatal("SetLogLevel must return the same builder")
	}
	if b.SetShowEntries([]string{"format=duration"}) != b {
		t.Fatal("SetShowEntries must return the same builder")
	}
	if b.SetOutputFormat("json") != b {
		t.Fatal("SetOutputFormat must return the same builder")
	}
	if b.SetInputFilePath("in.mp4") != b {
		t.Fatal("SetInputFilePath must return the same builder")
	}
	if b.SelectStreams("a:0") != b {
		t.Fatal("SelectStreams must return the same builder")
	}
}

func TestFfprobeBuilderBuildEmpty(t *testing.T) {
	got := NewFfprobeBuilder("ffprobe").Build()
	if len(got) != 0 {
		t.Fatalf("Build() = %v, want empty", got)
	}
}

func TestFfprobeBuilderBuildAllOptions(t *testing.T) {
	got := NewFfprobeBuilder("C:\\tools\\ffprobe.exe").
		SetLogLevel("error").
		SetShowEntries([]string{"stream=bit_rate", "format=duration", "format=size"}).
		SetOutputFormat("json").
		SetInputFilePath("in.mp4").
		SelectStreams("a:0").
		Build()
	want := []string{
		"-v", "error",
		"-show_entries", "stream=bit_rate",
		"-show_entries", "format=duration",
		"-show_entries", "format=size",
		"-of", "json",
		"in.mp4",
		"-select_streams", "a:0",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Build() = %v, want %v", got, want)
	}
}

func TestFfprobeBuilderBuildNoInput(t *testing.T) {
	got := NewFfprobeBuilder("ffprobe").
		SetLogLevel("error").
		SetOutputFormat("json").
		Build()
	want := []string{"-v", "error", "-of", "json"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Build() = %v, want %v", got, want)
	}
}