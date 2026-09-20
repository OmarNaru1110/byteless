package domain

import (
	"path/filepath"
	"testing"
)

func TestVideoFullPath(t *testing.T) {
	v := Video{Path: "movies", Name: "clip.mp4"}
	if got := v.FullPath(); got != filepath.Join("movies", "clip.mp4") {
		t.Fatalf("Video.FullPath() = %q, want %q", got, filepath.Join("movies", "clip.mp4"))
	}
}

func TestOutputVideoFullPath(t *testing.T) {
	o := OutputVideo{OutputPath: "out", OutputName: "result.mp4"}
	if got := o.FullPath(); got != filepath.Join("out", "result.mp4") {
		t.Fatalf("OutputVideo.FullPath() = %q, want %q", got, filepath.Join("out", "result.mp4"))
	}
}

func TestVideoEncoderValues(t *testing.T) {
	if H264 != VideoEncoder("libx264") {
		t.Fatalf("H264 = %q, want libx264", H264)
	}
	if H265 != VideoEncoder("libx265") {
		t.Fatalf("H265 = %q, want libx265", H265)
	}
}

func TestGetEncoders(t *testing.T) {
	encoders := GetEncoders()
	if len(encoders) != 2 {
		t.Fatalf("GetEncoders() returned %d encoders, want 2", len(encoders))
	}

	ids := make(map[VideoEncoder]bool)
	for i, e := range encoders {
		if ids[e.ID] {
			t.Fatalf("duplicate encoder id %q at index %d", e.ID, i)
		}
		ids[e.ID] = true
		if e.Name == "" || e.Description == "" {
			t.Fatalf("encoder %q missing name or description", e.ID)
		}
	}

	if encoders[0].ID != H264 {
		t.Fatalf("first encoder = %q, want %q", encoders[0].ID, H264)
	}
	if encoders[1].ID != H265 {
		t.Fatalf("second encoder = %q, want %q", encoders[1].ID, H265)
	}
}

func TestGetEncodersStableAcrossCalls(t *testing.T) {
	first := GetEncoders()
	second := GetEncoders()
	if len(first) != len(second) {
		t.Fatalf("encoder list length differs between calls: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("encoder list differs at index %d: %v vs %v", i, first[i], second[i])
		}
	}
}