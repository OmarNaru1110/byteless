package domain

import "path/filepath"

type Video struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Size         int    `json:"size"`         // bytes
	Duration     int    `json:"duration"`     // seconds
	AudioBitrate int    `json:"audioBitrate"` // kbps
}

func (v Video) FullPath() string {
	return filepath.Join(v.Path, v.Name)
}
