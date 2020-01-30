package sproxy

import (
	"io"
	"os"
	"time"
)

type Logger interface {
	Alert(s string) (err error)
	Crit(s string) (err error)
	Err(s string) (err error)
	Warning(s string) (err error)
	Notice(s string) (err error)
	Info(s string) (err error)
	Debug(s string) (err error)
}

var log Logger

type dumplogger struct {
}

func (l *dumplogger) Alert(s string) (err error) {
	return nil
}
func (l *dumplogger) Crit(s string) (err error) {
	return nil
}
func (l *dumplogger) Err(s string) (err error) {
	return nil
}
func (l *dumplogger) Warning(s string) (err error) {
	return nil
}
func (l *dumplogger) Notice(s string) (err error) {
	return nil
}
func (l *dumplogger) Info(s string) (err error) {
	return nil
}
func (l *dumplogger) Debug(s string) (err error) {
	return nil
}

type fileLogger struct {
	file io.Writer
}

func newFileLogger() (Logger, error) {
	filename := "sproxy.log"
	if f := os.Getenv("SPROXY_LOG_FILE"); f != "" {
		filename = f
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0)
	if err != nil {
		return nil, err
	}
	return &fileLogger{file}, nil
}

func (l *fileLogger) write(level string, s string) (err error) {
	now := time.Now().Format(time.RFC1123Z)
	_, err = l.file.Write([]byte(now + " " + level + " " + s + "\n"))
	return
}

func (l *fileLogger) Alert(s string) (err error) {
	return l.write("alert", s)
}
func (l *fileLogger) Crit(s string) (err error) {
	return l.write("crit", s)
}
func (l *fileLogger) Err(s string) (err error) {
	return l.write("error", s)
}
func (l *fileLogger) Warning(s string) (err error) {
	return l.write("warn", s)
}
func (l *fileLogger) Notice(s string) (err error) {
	return l.write("notice", s)
}
func (l *fileLogger) Info(s string) (err error) {
	return l.write("info", s)
}
func (l *fileLogger) Debug(s string) (err error) {
	return l.write("debug", s)
}
