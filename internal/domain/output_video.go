package domain

import "path/filepath"

type OutputVideo struct {
	TargetSize            int // bytes
	TargetSizeAfterMargin int // bytes
	OutputPath            string
	OutputName            string
	ElapsedTime           int // seconds
}

func (v OutputVideo) FullPath() string {
	return filepath.Join(v.OutputPath, v.OutputName)
}
