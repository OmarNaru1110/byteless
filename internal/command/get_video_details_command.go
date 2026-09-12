package command

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/OmarNaru1110/byteless/internal/builder"
)

type FFprobeVideoDetails struct {
	Streams []struct {
		BitRate string `json:"bit_rate"`
	} `json:"streams"`

	Format struct {
		Duration string `json:"duration"`
		Size     string `json:"size"`
	} `json:"format"`
}

type GetVideoDetailsCommand struct {
	builder *builder.FfprobeBuilder
}

func NewGetVideoDetailsCommand(inputFilePath string) (*GetVideoDetailsCommand, error) {
	if _, err := os.Stat(inputFilePath); err != nil {
		return nil, fmt.Errorf("NewGetVideoDetailsCommand: failed to stat input file: %w", err)
	}

	var ffprobePath string
	switch runtime.GOOS {
	case "windows":
		ffprobePath = "ffprobe.exe"
	case "darwin", "linux":
		ffprobePath = "ffprobe"
	default:
		return nil, fmt.Errorf("NewGetVideoDetailsCommand: unsupported operating system: %s", runtime.GOOS)
	}

	return &GetVideoDetailsCommand{
		builder: builder.NewFfprobeBuilder(ffprobePath).
			SetLogLevel("error").
			SelectStreams("a:0").
			SetShowEntries([]string{"stream=bit_rate", "format=duration", "format=size"}).
			SetOutputFormat("json").
			SetInputFilePath(inputFilePath),
	}, nil
}

func (c *GetVideoDetailsCommand) Execute() (*FFprobeVideoDetails, error) {
	if c.builder == nil {
		err := fmt.Errorf("FfprobeBuilder is nil")
		log.Printf("GetVideoDetailsCommand: Builder validation failed: %v", err)
		return nil, err
	}

	ffprobePath := c.builder.GetFfprobePath()

	cmdArgs := c.builder.Build()
	if len(cmdArgs) == 0 {
		return nil, fmt.Errorf("GetVideoDetailsCommand: no arguments provided")
	}

	log.Printf("GetVideoDetailsCommand: executing command: %s %s", ffprobePath, strings.Join(cmdArgs, " "))

	cmd := exec.Command(ffprobePath, cmdArgs...)

	var stderr strings.Builder
	cmd.Stderr = &stderr

	stdoutBytes, err := cmd.Output()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			log.Printf("GetVideoDetailsCommand: stderr: %s", msg)
		}
		return nil, fmt.Errorf("GetVideoDetailsCommand: command failed: %w", err)
	}

	stdoutStr := strings.TrimSpace(string(stdoutBytes))
	if stdoutStr == "" {
		return nil, fmt.Errorf("GetVideoDetailsCommand: no output from command")
	}

	var details FFprobeVideoDetails
	if err := json.Unmarshal(stdoutBytes, &details); err != nil {
		return nil, fmt.Errorf("GetVideoDetailsCommand: failed to parse output %q: %w", stdoutStr, err)
	}

	log.Printf("GetVideoDetailsCommand: result = duration %s, size %s", details.Format.Duration, details.Format.Size)
	return &details, nil
}