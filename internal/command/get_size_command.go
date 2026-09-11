package command

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/OmarNaru1110/byteless/internal/builder"
)

type GetVideoSizeCommand struct {
	builder *builder.FfprobeBuilder
}

func NewGetVideoSizeInBytesCommand(inputFilePath string) (*GetVideoSizeCommand, error) {
	_, err := os.Stat(inputFilePath)
	if err != nil {
		log.Fatalf("Failed to stat input file: %v", err)
		return nil, err
	}

	var ffprobePath string
	switch os := runtime.GOOS; os {
	case "windows":
		ffprobePath = "ffprobe.exe"
	case "darwin", "linux":
		ffprobePath = "ffprobe"
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", os)
	}

	return &GetVideoSizeCommand{
		builder: builder.NewFfprobeBuilder(ffprobePath).
			SetLogLevel("error").
			SetShowEntries("format=size").
			SetOutputFormat("default=noprint_wrappers=1:nokey=1").
			SetInputFilePath(inputFilePath),
	}, nil
}

func (c *GetVideoSizeCommand) Execute() (int64, error) {
	if c.builder == nil {
		err := fmt.Errorf("FfprobeBuilder is nil")
		log.Printf("NewGetVideoSizeInBytesCommand: Builder validation failed: %v", err)
		return 0, err
	}

	ffprobePath := c.builder.GetFfprobePath()

	cmdArgs := c.builder.Build()

	if len(cmdArgs) == 0 {
		return 0, fmt.Errorf("NewGetVideoSizeInBytesCommand: no arguments provided")
	}

	log.Printf("NewGetVideoSizeInBytesCommand: executing command: %s %s", ffprobePath, strings.Join(cmdArgs, " "))

	cmd := exec.Command(ffprobePath, cmdArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, fmt.Errorf("NewGetVideoSizeInBytesCommand: failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, fmt.Errorf("NewGetVideoSizeInBytesCommand: failed to create stderr pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("NewGetVideoSizeInBytesCommand: failed to start command: %w", err)
	}

	var output strings.Builder
	scanner := bufio.NewScanner(io.TeeReader(stderr, &output))
	for scanner.Scan() {
		log.Printf("NewGetVideoSizeInBytesCommand: stderr: %s", scanner.Text())
	}

	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		_ = cmd.Wait()
		return 0, fmt.Errorf("NewGetVideoSizeInBytesCommand: failed to read stdout: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return 0, fmt.Errorf("NewGetVideoSizeInBytesCommand: command failed: %w", err)
	}

	stdoutStr := strings.TrimSpace(string(stdoutBytes))
	if stdoutStr == "" {
		return 0, fmt.Errorf("NewGetVideoSizeInBytesCommand: no output from command")
	}

	bytes, err := strconv.ParseInt(stdoutStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("NewGetVideoSizeInBytesCommand: failed to parse output %q: %w", stdoutStr, err)
	}

	log.Printf("NewGetVideoSizeInBytesCommand: result = %d bytes", bytes)
	return bytes, nil
}
