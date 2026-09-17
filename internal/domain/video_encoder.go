package domain

type VideoEncoder string

const (
	H264 VideoEncoder = "libx264"
	H265 VideoEncoder = "libx265"
)

type EncoderInfo struct {
	ID          VideoEncoder `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
}

var availableEncoders = []EncoderInfo{
	{ID: H264, Name: "H.264", Description: "Fast, widely compatible"},
	{ID: H265, Name: "H.265", Description: "Better compression, slower"},
}

func GetEncoders() []EncoderInfo {
	return availableEncoders
}
