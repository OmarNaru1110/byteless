package domain

type OutputVideo struct {
	TargetSize            int // bytes
	TargetSizeAfterMargin int // bytes
	OutputPath            string
	OutputName            string
	ElapsedTime           int // seconds
}
