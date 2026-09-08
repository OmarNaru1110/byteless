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

type GetBytesCommand struct {
	builder *builder.FfprobeBuilder
}

func NewGetBytesCommand(inputFilePath string) (*GetBytesCommand, error) {
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

	return &GetBytesCommand{
		builder: builder.NewFfprobeBuilder(ffprobePath).
			SetLogLevel("error").
			SetShowEntries("format=size").
			SetOutputFormat("default=noprint_wrappers=1:nokey=1").
			SetInputFilePath(inputFilePath),
	}, nil
}

func (c *GetBytesCommand) Execute() (int64, error) {
	if c.builder == nil {
		err := fmt.Errorf("FfprobeBuilder is nil")
		log.Printf("GetBytesCommand: Builder validation failed: %v", err)
		return 0, err
	}

	ffprobePath := c.builder.GetFfprobePath()

	cmdArgs := c.builder.Build()

	if len(cmdArgs) == 0 {
		return 0, fmt.Errorf("GetBytesCommand: no arguments provided")
	}

	log.Printf("GetBytesCommand: executing command: %s %s", ffprobePath, strings.Join(cmdArgs, " "))

	cmd := exec.Command(ffprobePath, cmdArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, fmt.Errorf("GetBytesCommand: failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, fmt.Errorf("GetBytesCommand: failed to create stderr pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("GetBytesCommand: failed to start command: %w", err)
	}

	var output strings.Builder
	scanner := bufio.NewScanner(io.TeeReader(stderr, &output))
	for scanner.Scan() {
		log.Printf("GetBytesCommand: stderr: %s", scanner.Text())
	}

	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		_ = cmd.Wait()
		return 0, fmt.Errorf("GetBytesCommand: failed to read stdout: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return 0, fmt.Errorf("GetBytesCommand: command failed: %w", err)
	}

	stdoutStr := strings.TrimSpace(string(stdoutBytes))
	if stdoutStr == "" {
		return 0, fmt.Errorf("GetBytesCommand: no output from command")
	}

	bytes, err := strconv.ParseInt(stdoutStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("GetBytesCommand: failed to parse output %q: %w", stdoutStr, err)
	}

	log.Printf("GetBytesCommand: result = %d bytes", bytes)
	return bytes, nil
}
