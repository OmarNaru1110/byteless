package domain

type Video struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Size         int    `json:"size"`         // bytes
	Duration     int    `json:"duration"`     // seconds
	AudioBitrate int    `json:"audioBitrate"` // kbps
}
