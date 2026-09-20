package util

import (
	"regexp"
	"testing"
)

func TestConvertMBToBytes(t *testing.T) {
	tests := []struct {
		name string
		mb   float64
		want int
	}{
		{"zero", 0, 0},
		{"one", 1, 1048576},
		{"half", 0.5, 524288},
		{"one and a half", 1.5, 1572864},
		{"two and a half", 2.5, 2621440},
		{"negative", -2, -2097152},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertMBToBytes(tt.mb); got != tt.want {
				t.Fatalf("ConvertMBToBytes(%v) = %d, want %d", tt.mb, got, tt.want)
			}
		})
	}
}

func TestConvertBytesToMB(t *testing.T) {
	tests := []struct {
		name string
		bytes int
		want float64
	}{
		{"zero", 0, 0},
		{"one", 1048576, 1},
		{"one and a half", 1572864, 1.5},
		{"ten", 10485760, 10},
		{"negative", -1048576, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertBytesToMB(tt.bytes); got != tt.want {
				t.Fatalf("ConvertBytesToMB(%d) = %v, want %v", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestConvertRoundTrip(t *testing.T) {
	for _, b := range []int{0, 1048576, 1572864, 10 * 1024 * 1024, 333 * 1024 * 1024} {
		if got := ConvertMBToBytes(ConvertBytesToMB(b)); got != b {
			t.Fatalf("round trip failed for %d: got %d", b, got)
		}
	}
}

func TestCalculateBitrate(t *testing.T) {
	tests := []struct {
		name         string
		videoSize    int
		durationSecs int
		audioBitrate int
		want         int
	}{
		{"10MB over 10s with 128k audio", 10 * 1024 * 1024, 10, 128, 8064},
		{"1MB over 1s with 128k audio", 1024 * 1024, 1, 128, 8064},
		{"5MB over 10s with 64k audio", 5 * 1024 * 1024, 10, 64, 4032},
		{"zero size", 0, 10, 0, 0},
		{"no audio", 10485760, 1, 0, 81920},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateBitrate(tt.videoSize, tt.durationSecs, tt.audioBitrate); got != tt.want {
				t.Fatalf("CalculateBitrate(%d, %d, %d) = %d, want %d", tt.videoSize, tt.durationSecs, tt.audioBitrate, got, tt.want)
			}
		})
	}
}

func TestApplyMargin(t *testing.T) {
	tests := []struct {
		name   string
		value  float64
		margin float64
		want   float64
	}{
		{"ten percent of 100", 100, 10, 90},
		{"ten percent of 50", 50, 10, 45},
		{"zero value", 0, 10, 0},
		{"zero margin", 100, 0, 100},
		{"negative margin increases", 100, -10, 110},
		{"fractional", 12.5, 20, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ApplyMargin(tt.value, tt.margin); got != tt.want {
				t.Fatalf("ApplyMargin(%v, %v) = %v, want %v", tt.value, tt.margin, got, tt.want)
			}
		})
	}
}

func TestGenerateName(t *testing.T) {
	for _, ext := range []string{"mp4", "mkv", "webm"} {
		name := GenerateName(ext)
		pattern := `^byteless_\d+\.` + ext + `$`
		if !regexp.MustCompile(pattern).MatchString(name) {
			t.Fatalf("GenerateName(%q) = %q, want match %q", ext, name, pattern)
		}
	}
}

func TestParseTimeToSeconds(t *testing.T) {
	tests := []struct {
		name string
		h, m, s, cs string
		want float64
	}{
		{"simple timestamp", "1", "2", "3", "0", 3723},
		{"with centiseconds", "1", "2", "3", "50", 3723.5},
		{"all zeros", "0", "0", "0", "0", 0},
		{"minutes only", "0", "1", "30", "00", 90},
		{"invalid parts become zero", "a", "b", "c", "d", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTimeToSeconds(tt.h, tt.m, tt.s, tt.cs); got != tt.want {
				t.Fatalf("ParseTimeToSeconds(%q, %q, %q, %q) = %v, want %v", tt.h, tt.m, tt.s, tt.cs, got, tt.want)
			}
		})
	}
}

func TestConvertMicrosecondsToSeconds(t *testing.T) {
	tests := []struct {
		name string
		us   int64
		want float64
	}{
		{"zero", 0, 0},
		{"half second", 500000, 0.5},
		{"one second", 1000000, 1},
		{"negative", -500000, -0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertMicrosecondsToSeconds(tt.us); got != tt.want {
				t.Fatalf("ConvertMicrosecondsToSeconds(%d) = %v, want %v", tt.us, got, tt.want)
			}
		})
	}
}