package testhelper

import (
	"os"
	"strings"
)

const (
	MediaHelperEnvKey     = "BYTELESS_MEDIA_HELPER"
	MediaHelperFailEnvKey = "BYTELESS_MEDIA_HELPER_FAIL"
)

func FakeMediaBinary() string {
	if p, err := os.Executable(); err == nil {
		return p
	}
	return os.Args[0]
}

func IsMediaHelper() bool {
	return os.Getenv(MediaHelperEnvKey) == "1"
}

func SetMediaHelperEnv() {
	os.Setenv(MediaHelperEnvKey, "1")
}

func HelperMain() {
	args := ""
	if len(os.Args) > 1 {
		args = strings.Join(os.Args[1:], " ")
	}

	if os.Getenv(MediaHelperFailEnvKey) == "1" {
		os.Stderr.WriteString("fake media binary failure\n")
		os.Exit(1)
	}

	if strings.Contains(args, "-progress") {
		os.Stdout.WriteString("frame=5\n" +
			"fps=25.00\n" +
			"out_time_us=500000\n" +
			"out_time_ms=500\n" +
			"out_time=00:00:00.500000\n" +
			"progress=continue\n")
		os.Exit(0)
	}

	os.Stdout.WriteString(`{"streams":[{"index":0,"bit_rate":"128000"}],"format":{"duration":"100.000000","size":"104857600"}}`)
	os.Exit(0)
}