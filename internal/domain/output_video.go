package domain

import "path/filepath"

type OutputVideo struct {
	TargetSize            float64      `json:"targetSize"`            // MB
	TargetSizeAfterMargin float64      `json:"targetSizeAfterMargin"` // MB
	OutputPath            string       `json:"outputPath"`
	OutputName            string       `json:"outputName"`
	Size                  int          `json:"size"`        // bytes
	ElapsedTime           int          `json:"elapsedTime"` // seconds
	Encoder               VideoEncoder `json:"encoder"`
}

func (v OutputVideo) FullPath() string {
	return filepath.Join(v.OutputPath, v.OutputName)
}
