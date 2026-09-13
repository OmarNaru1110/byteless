package util

import (
	"fmt"
	"time"
)

func ConvertMBToBytes(mb int) int {
	return mb * 1024 * 1024
}

func ConvertBytesToMB(bytes int) int {
	return bytes / (1024 * 1024)
}

// CalculateBitrate calculates the bitrate of a video in kbps, given the video size in bytes, duration in seconds, and audio bitrate in kbps.
func CalculateBitrate(videoSizeBytes int, durationSeconds int, audioBitrate int) int {
	bitrate := (ConvertBytesToMB(videoSizeBytes)*8192)/durationSeconds - audioBitrate
	return bitrate
}

func ApplyMargin(value int, marginPercent float64) int {
	margin := float64(value) * (marginPercent / 100)
	return int(float64(value) - margin)
}

func GenerateName(extension string) string {
	return fmt.Sprintf("byteless_%d.%s", time.Now().Unix(), extension)
}
