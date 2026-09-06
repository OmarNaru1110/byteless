package builder

import (
	"log"
	"strconv"
)

type ffmpegOptions struct {
	inputFilePath  string
	outputFilePath string
	videoCodec     string
	videoBitrate   string
	pass           int
	audioCodec     string
	audioBitrate   string
	disableAudio   bool
	format         string
}

type FfmpegBuilder struct {
	ffmpegPath string
	options    ffmpegOptions
}

func NewFfmpegBuilder(ffmpegPath string) *FfmpegBuilder {
	return &FfmpegBuilder{
		ffmpegPath: ffmpegPath,
	}
}

func (b *FfmpegBuilder) SetInputFilePath(inputFilePath string) *FfmpegBuilder {
	b.options.inputFilePath = inputFilePath
	log.Printf("FfmpegBuilder: input file path set to %q", inputFilePath)
	return b
}

func (b *FfmpegBuilder) SetOutputFilePath(outputFilePath string) *FfmpegBuilder {
	b.options.outputFilePath = outputFilePath
	log.Printf("FfmpegBuilder: output file path set to %q", outputFilePath)
	return b
}

func (b *FfmpegBuilder) SetVideoCodec(videoCodec string) *FfmpegBuilder {
	b.options.videoCodec = videoCodec
	log.Printf("FfmpegBuilder: video codec set to %q", videoCodec)
	return b
}

func (b *FfmpegBuilder) SetVideoBitrate(videoBitrate string) *FfmpegBuilder {
	b.options.videoBitrate = videoBitrate
	log.Printf("FfmpegBuilder: video bitrate set to %q", videoBitrate)
	return b
}

func (b *FfmpegBuilder) SetPass(pass int) *FfmpegBuilder {
	b.options.pass = pass
	log.Printf("FfmpegBuilder: pass set to %d", pass)
	return b
}

func (b *FfmpegBuilder) SetAudioCodec(audioCodec string) *FfmpegBuilder {
	b.options.audioCodec = audioCodec
	log.Printf("FfmpegBuilder: audio codec set to %q", audioCodec)
	return b
}

func (b *FfmpegBuilder) SetAudioBitrate(audioBitrate string) *FfmpegBuilder {
	b.options.audioBitrate = audioBitrate
	log.Printf("FfmpegBuilder: audio bitrate set to %q", audioBitrate)
	return b
}

func (b *FfmpegBuilder) DisableAudio() *FfmpegBuilder {
	b.options.disableAudio = true
	log.Println("FfmpegBuilder: audio disabled")
	return b
}

func (b *FfmpegBuilder) SetFormat(format string) *FfmpegBuilder {
	b.options.format = format
	log.Printf("FfmpegBuilder: format set to %q", format)
	return b
}

func (b *FfmpegBuilder) Build() []string {
	var args []string
	log.Printf("FfmpegBuilder: building args with options: %+v", b.options)

	if b.options.inputFilePath != "" {
		args = append(args, "-i", b.options.inputFilePath)
	}

	if b.options.outputFilePath != "" {
		args = append(args, b.options.outputFilePath)
	} else {
		args = append(args, "NUL")
	}

	if b.options.videoCodec != "" {
		args = append(args, "-c:v", b.options.videoCodec)
	}

	if b.options.videoBitrate != "" {
		args = append(args, "-b:v", b.options.videoBitrate)
	}

	if b.options.pass > 0 {
		args = append(args, "-pass", strconv.Itoa(b.options.pass))
	}

	if b.options.audioCodec != "" {
		args = append(args, "-c:a", b.options.audioCodec)
	}

	if b.options.audioBitrate != "" {
		args = append(args, "-b:a", b.options.audioBitrate)
	}

	if b.options.disableAudio {
		args = append(args, "-an")
	}

	if b.options.format != "" {
		args = append(args, "-f", b.options.format)
	}

	log.Printf("FfmpegBuilder: built args: %v", args)
	return args
}
