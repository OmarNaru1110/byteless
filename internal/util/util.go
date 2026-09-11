package util

func ConvertMBToBytes(mb int) int {
	return mb * 1024 * 1024
}

func ConvertBytesToMB(bytes int) int {
	return bytes / (1024 * 1024)
}

func CalculateBitrate(videoSizeBytes int, durationSeconds int, audioBitrate int) int {
	bitrate := (ConvertBytesToMB(videoSizeBytes)*8192)/durationSeconds - audioBitrate
	return bitrate
}
