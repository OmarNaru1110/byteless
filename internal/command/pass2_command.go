package command

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/OmarNaru1110/byteless/internal/builder"
	"github.com/OmarNaru1110/byteless/internal/domain"
	"github.com/OmarNaru1110/byteless/internal/util"
)

type TwoPassEncodePass2Command struct {
	builder *builder.FfmpegBuilder
}

func NewTwoPassEncodePass2Command(ffmpegPath string, inputPath string, targetVideoBitrateKbps int, outputFilePath string, encoder domain.VideoEncoder) *TwoPassEncodePass2Command {
	return &TwoPassEncodePass2Command{
		builder: builder.NewFfmpegBuilder(ffmpegPath).
			SetInputFilePath(inputPath).
			SetVideoCodec(string(encoder)).
			SetVideoBitrate(fmt.Sprintf("%dk", targetVideoBitrateKbps)).
			SetPass(2).
			SetAudioCodec("aac").
			SetAudioBitrate("128k").
			SetProgressPipe("pipe:1").
			SetOutputFilePath(outputFilePath),
	}
}

func (c *TwoPassEncodePass2Command) Execute(ctx context.Context, totalSeconds int) error {
	if c.builder == nil {
		err := fmt.Errorf("FfmpegBuilder is nil")
		log.Printf("TwoPassEncodePass2Command: Builder validation failed: %v", err)
		return err
	}

	ffmpegPath := c.builder.GetFfmpegPath()

	cmdArgs := c.builder.Build()

	if len(cmdArgs) == 0 {
		return fmt.Errorf("TwoPassEncodePass2Command: no arguments provided")
	}

	log.Printf("TwoPassEncodePass2Command: executing %s %s", ffmpegPath, strings.Join(cmdArgs, " "))

	cmd := exec.CommandContext(ctx, ffmpegPath, cmdArgs...)
	util.HideWindow(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("TwoPassEncodePass2Command: failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("TwoPassEncodePass2Command: failed to start ffmpeg: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "out_time_us=") {
			timeStr := strings.TrimPrefix(line, "out_time_us=")
			timeUs, err := strconv.Atoi(timeStr)
			if err != nil {
				log.Printf("TwoPassEncodePass2Command: failed to parse out_time_us: %v", err)
				continue
			}

			timeSeconds := util.ConvertMicrosecondsToSeconds(int64(timeUs))
			percent := float64(timeSeconds) / float64(totalSeconds) * 100
			fmt.Printf("\rTwoPassEncodePass2Command: encoding progress: %.2f%%", percent)

			EmitEvent(ctx, "pass2Progress", fmt.Sprintf("%.0f", percent))
		}
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println()
		return fmt.Errorf("TwoPassEncodePass2Command: ffmpeg exited with error: %w", err)
	}

	fmt.Println()
	log.Println("TwoPassEncodePass2Command: encoding completed successfully")
	return nil
}
