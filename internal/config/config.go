package config

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Config holds the paths to the external tools the app shells out to.
type Config struct {
	FFmpegPath  string
	FFprobePath string
}

const (
	// FFmpegPathEnv overrides where ffmpeg is resolved from.
	FFmpegPathEnv = "BYTELESS_FFMPEG_PATH"
	// FFprobePathEnv overrides where ffprobe is resolved from.
	FFprobePathEnv = "BYTELESS_FFPROBE_PATH"
)

// Default returns a Config with ffmpeg and ffprobe resolved automatically.
func Default() Config {
	return Config{
		FFmpegPath:  ResolveBinary("ffmpeg", FFmpegPathEnv),
		FFprobePath: ResolveBinary("ffprobe", FFprobePathEnv),
	}
}

// ResolveBinary locates binaryName. Order of preference:
//  1. the path set via envOverride, if non-empty
//  2. a bundled copy next to the running executable or in its bin/ subfolder
//  3. a copy under the working directory or build/bin
//  4. the system PATH
//
// As a last resort the bare binary name is returned so callers still get a
// useful error from exec instead of a missing command.
func ResolveBinary(binaryName, envOverride string) string {
	if envOverride != "" {
		if p := os.Getenv(envOverride); p != "" {
			log.Printf("Config: using %s from env %s=%q", binaryName, envOverride, p)
			return p
		}
	}

	binary := binaryName + binarySuffix()
	for _, dir := range candidateDirs() {
		p := filepath.Join(dir, binary)
		if isFile(p) {
			log.Printf("Config: using bundled %s at %q", binaryName, p)
			return p
		}
	}

	if p, err := exec.LookPath(binaryName); err == nil {
		log.Printf("Config: using %s from PATH at %q", binaryName, p)
		return p
	}

	log.Printf("Config: %s not found, falling back to bare %q", binaryName, binaryName)
	return binaryName
}

func binarySuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func candidateDirs() []string {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe), filepath.Join(filepath.Dir(exe), "bin"))
	}
	if wd, err := os.Getwd(); err == nil {
		dirs = append(dirs, wd, filepath.Join(wd, "build", "bin"), filepath.Join(wd, "bin"))
	}
	return dirs
}

func isFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular()
}