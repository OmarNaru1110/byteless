package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const uniqueBinName = "byteless-test-unique-binary"

func TestBinarySuffix(t *testing.T) {
	if runtime.GOOS == "windows" {
		if got := binarySuffix(); got != ".exe" {
			t.Fatalf("binarySuffix() = %q, want .exe", got)
		}
	} else {
		if got := binarySuffix(); got != "" {
			t.Fatalf("binarySuffix() = %q, want empty", got)
		}
	}
}

func TestIsFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if !isFile(file) {
		t.Fatalf("isFile(%q) = false, want true", file)
	}
	if isFile(dir) {
		t.Fatalf("isFile(%q) = true, want false for directory", dir)
	}
	if isFile(filepath.Join(dir, "missing")) {
		t.Fatalf("isFile(%q) = true, want false for missing path", filepath.Join(dir, "missing"))
	}
}

func TestResolveBinaryFromEnv(t *testing.T) {
	want := filepath.Join("C:", "tools", "fake.exe")
	t.Setenv(uniqueBinName, want)
	if got := ResolveBinary(uniqueBinName, uniqueBinName); got != want {
		t.Fatalf("ResolveBinary() = %q, want %q", got, want)
	}
}

func TestResolveBinaryIgnoresEmptyEnv(t *testing.T) {
	t.Setenv(uniqueBinName, "")
	if got := ResolveBinary(uniqueBinName, uniqueBinName); got != uniqueBinName {
		t.Fatalf("ResolveBinary() = %q, want the bare name %q", got, uniqueBinName)
	}
}

func TestResolveBinaryBundledInConfigDir(t *testing.T) {
	cfgRoot := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", cfgRoot)
	} else {
		t.Setenv("XDG_CONFIG_HOME", cfgRoot)
	}

	binDir := filepath.Join(cfgRoot, "byteless")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(binDir, uniqueBinName+binarySuffix())
	if err := os.WriteFile(want, []byte("x"), 0755); err != nil {
		t.Fatal(err)
	}

	if got := ResolveBinary(uniqueBinName, ""); got != want {
		t.Fatalf("ResolveBinary() = %q, want bundled path %q", got, want)
	}
}

func TestResolveBinaryFallsBackToBareName(t *testing.T) {
	t.Setenv(uniqueBinName, "")
	if got := ResolveBinary(uniqueBinName, ""); got != uniqueBinName {
		t.Fatalf("ResolveBinary() = %q, want bare name %q", got, uniqueBinName)
	}
}

func TestDefaultUsesEnvOverrides(t *testing.T) {
	t.Setenv(FFmpegPathEnv, "fake-ffmpeg")
	t.Setenv(FFprobePathEnv, "fake-ffprobe")
	c := Default()
	if c.FFmpegPath != "fake-ffmpeg" {
		t.Fatalf("Default().FFmpegPath = %q, want fake-ffmpeg", c.FFmpegPath)
	}
	if c.FFprobePath != "fake-ffprobe" {
		t.Fatalf("Default().FFprobePath = %q, want fake-ffprobe", c.FFprobePath)
	}
}

func TestCandidateDirs(t *testing.T) {
	dirs := candidateDirs()
	if len(dirs) == 0 {
		t.Fatal("candidateDirs() returned empty")
	}
	foundConfigDir := false
	for _, d := range dirs {
		if filepath.Base(d) == "byteless" {
			foundConfigDir = true
			break
		}
	}
	if !foundConfigDir {
		t.Fatalf("candidateDirs() %v does not include the byteless config dir", dirs)
	}
}