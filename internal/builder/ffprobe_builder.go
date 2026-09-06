package builder

type FfprobeOptions struct {
	logLevel      string
	showEntries   string
	outputFormat  string
	inputFilePath string
}

type FfprobeBuilder struct {
	ffprobePath string
	options     FfprobeOptions
}

func NewFfprobeBuilder(ffprobePath string) *FfprobeBuilder {
	return &FfprobeBuilder{
		ffprobePath: ffprobePath,
	}
}

func (b *FfprobeBuilder) SetLogLevel(logLevel string) *FfprobeBuilder {
	b.options.logLevel = logLevel
	return b
}

func (b *FfprobeBuilder) SetShowEntries(showEntries string) *FfprobeBuilder {
	b.options.showEntries = showEntries
	return b
}

func (b *FfprobeBuilder) SetOutputFormat(outputFormat string) *FfprobeBuilder {
	b.options.outputFormat = outputFormat
	return b
}

func (b *FfprobeBuilder) SetInputFilePath(inputFilePath string) *FfprobeBuilder {
	b.options.inputFilePath = inputFilePath
	return b
}

func (b *FfprobeBuilder) Build() []string {
	args := []string{}
	if b.options.logLevel != "" {
		args = append(args, "-v", b.options.logLevel)
	}
	if b.options.showEntries != "" {
		args = append(args, "-show_entries", b.options.showEntries)
	}
	if b.options.outputFormat != "" {
		args = append(args, "-of", b.options.outputFormat)
	}
	if b.options.inputFilePath != "" {
		args = append(args, b.options.inputFilePath)
	}
	return args
}
