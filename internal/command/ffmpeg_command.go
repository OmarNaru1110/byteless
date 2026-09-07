package command

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/OmarNaru1110/byteless/internal/builder"
)

type FfmpegCommand struct {
	Builder *builder.FfmpegBuilder
}

var durationRe = regexp.MustCompile(`Duration:\s*(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
var timeRe = regexp.MustCompile(`time=\s*(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)

func parseTimeToSeconds(h, m, s, cs string) float64 {
	hours, _ := strconv.ParseFloat(h, 64)
	minutes, _ := strconv.ParseFloat(m, 64)
	seconds, _ := strconv.ParseFloat(s, 64)
	centiseconds, _ := strconv.ParseFloat(cs, 64)
	return hours*3600 + minutes*60 + seconds + centiseconds/100
}

func (c *FfmpegCommand) Execute(args any) (any, error) {
	if c.Builder == nil {
		err := fmt.Errorf("FfmpegBuilder is nil")
		log.Printf("FfmpegCommand: Builder validation failed: %v", err)
		return nil, err
	}

	ffmpegPath := c.Builder.GetFfmpegPath()

	var cmdArgs []string
	switch a := args.(type) {
	case []string:
		cmdArgs = a
	default:
		return nil, fmt.Errorf("FfmpegCommand: unsupported args type %T, expected []string", args)
	}

	if len(cmdArgs) == 0 {
		return nil, fmt.Errorf("FfmpegCommand: no arguments provided")
	}

	log.Printf("FfmpegCommand: executing %s %s", ffmpegPath, strings.Join(cmdArgs, " "))

	cmd := exec.Command(ffmpegPath, cmdArgs...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("FfmpegCommand: failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("FfmpegCommand: failed to start ffmpeg: %w", err)
	}

	var totalSeconds float64
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()

		if totalSeconds == 0 {
			if m := durationRe.FindStringSubmatch(line); m != nil {
				totalSeconds = parseTimeToSeconds(m[1], m[2], m[3], m[4])
				log.Printf("FfmpegCommand: detected duration %.2fs", totalSeconds)
			}
		}

		if strings.Contains(line, "time=") {
			if m := timeRe.FindStringSubmatch(line); m != nil {
				currentSeconds := parseTimeToSeconds(m[1], m[2], m[3], m[4])
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
		return nil, fmt.Errorf("FfmpegCommand: ffmpeg exited with error: %w", err)
	}

	fmt.Println()
	log.Println("FfmpegCommand: encoding completed successfully")
	return nil, nil
}
