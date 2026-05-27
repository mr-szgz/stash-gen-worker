package worker

type ffmpegConfig struct{}

func (c *ffmpegConfig) GetTranscodeInputArgs() []string {
	return nil
}

func (c *ffmpegConfig) GetTranscodeOutputArgs() []string {
	return nil
}
