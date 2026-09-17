package util

import (
	"fmt"
	"strconv"
	"time"
)

func ConvertMBToBytes(mb float64) int {
	return int(mb * 1024 * 1024)
}

func ConvertBytesToMB(bytes int) float64 {
	return float64(bytes) / (1024 * 1024)
}

// CalculateBitrate calculates the bitrate of a video in kbps, given the video size in bytes, duration in seconds, and audio bitrate in kbps.
func CalculateBitrate(videoSizeBytes int, durationSeconds int, audioBitrate int) int {
	bitrate := (ConvertBytesToMB(videoSizeBytes)*8192)/float64(durationSeconds) - float64(audioBitrate)
	return int(bitrate)
}

func ApplyMargin(value float64, marginPercent float64) float64 {
	margin := value * (marginPercent / 100)
	return value - margin
}

func GenerateName(extension string) string {
	return fmt.Sprintf("byteless_%d.%s", time.Now().Unix(), extension)
}

func ParseTimeToSeconds(h, m, s, cs string) float64 {
	hours, _ := strconv.ParseFloat(h, 64)
	minutes, _ := strconv.ParseFloat(m, 64)
	seconds, _ := strconv.ParseFloat(s, 64)
	centiseconds, _ := strconv.ParseFloat(cs, 64)
	return hours*3600 + minutes*60 + seconds + centiseconds/100
}

func ConvertMicrosecondsToSeconds(microseconds int64) float64 {
	return float64(microseconds) / 1_000_000
}
