package domain

import "path/filepath"

type OutputVideo struct {
	TargetSize            int          `json:"targetSize"`            // bytes
	TargetSizeAfterMargin int          `json:"targetSizeAfterMargin"` // bytes
	OutputPath            string       `json:"outputPath"`
	OutputName            string       `json:"outputName"`
	Size                  int          `json:"size"`        // bytes
	ElapsedTime           int          `json:"elapsedTime"` // seconds
	Encoder               VideoEncoder `json:"encoder"`
}

func (v OutputVideo) FullPath() string {
	return filepath.Join(v.OutputPath, v.OutputName)
}
