package sproxy

func NewLogger() (Logger, error) {
	return newFileLogger()
}
