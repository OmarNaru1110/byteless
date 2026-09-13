package command

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/OmarNaru1110/byteless/internal/builder"
	"github.com/OmarNaru1110/byteless/internal/util"
)

type TwoPassEncodePass1Command struct {
	builder *builder.FfmpegBuilder
}

func NewTwoPassEncodePass1Command(inputPath string, videoBitrateKbps int) *TwoPassEncodePass1Command {
	return &TwoPassEncodePass1Command{
		builder: builder.NewFfmpegBuilder("ffmpeg").
			SetInputFilePath(inputPath).
			SetVideoCodec("libx264").
			SetVideoBitrate(fmt.Sprintf("%dk", videoBitrateKbps)).
			SetPass(1).
			DisableAudio().
			SetFormat("null"),
	}
}

var durationRe = regexp.MustCompile(`Duration:\s*(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
var timeRe = regexp.MustCompile(`time=\s*(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)

func (c *TwoPassEncodePass1Command) Execute() error {
	if c.builder == nil {
		err := fmt.Errorf("FfmpegBuilder is nil")
		log.Printf("TwoPassEncodePass1Command: Builder validation failed: %v", err)
		return err
	}

	ffmpegPath := c.builder.GetFfmpegPath()

	cmdArgs := c.builder.Build()

	if len(cmdArgs) == 0 {
		return fmt.Errorf("TwoPassEncodePass1Command: no arguments provided")
	}

	log.Printf("TwoPassEncodePass1Command: executing %s %s", ffmpegPath, strings.Join(cmdArgs, " "))

	cmd := exec.Command(ffmpegPath, cmdArgs...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("TwoPassEncodePass1Command: failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("TwoPassEncodePass1Command: failed to start ffmpeg: %w", err)
	}

	var totalSeconds float64
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()

		if totalSeconds == 0 {
			if m := durationRe.FindStringSubmatch(line); m != nil {
				totalSeconds = util.ParseTimeToSeconds(m[1], m[2], m[3], m[4])
				log.Printf("TwoPassEncodePass1Command: detected duration %.2fs", totalSeconds)
			}
		}

		if strings.Contains(line, "time=") {
			if m := timeRe.FindStringSubmatch(line); m != nil {
				currentSeconds := util.ParseTimeToSeconds(m[1], m[2], m[3], m[4])
				if totalSeconds > 0 {
					pct := (currentSeconds / totalSeconds) * 100
					fmt.Printf("\r[%6.1f%%] %s", pct, strings.TrimSpace(line))
				} else {
					fmt.Printf("\r%s", strings.TrimSpace(line))
				}
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println()
		return fmt.Errorf("TwoPassEncodePass1Command: ffmpeg exited with error: %w", err)
	}

	fmt.Println()
	log.Println("TwoPassEncodePass1Command: encoding completed successfully")
	return nil
}
