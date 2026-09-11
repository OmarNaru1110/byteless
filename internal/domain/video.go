package domain

type Video struct {
	Path         string
	Name         string
	Size         int // bytes
	Duration     int // seconds
	AudioBitrate int // kbps
}
