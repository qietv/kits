package utils

import "github.com/hanskorg/logkit"

var (
	logger Logger
)

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

type SimpleLogger struct{}

func NewSimpleLogger() Logger {
	logger = &SimpleLogger{}
	return logger
}
func SetLogger(l Logger) {
	logger = l
}

func GetLogger() Logger {
	return logger
}

func (s *SimpleLogger) Debug(format string, args ...interface{}) {
	logkit.Debugf(format, args...)
}

func (s *SimpleLogger) Info(format string, args ...interface{}) {
	logkit.Infof(format, args...)
}

func (s *SimpleLogger) Warn(format string, args ...interface{}) {
	logkit.Warnf(format, args...)
}

func (s *SimpleLogger) Error(format string, args ...interface{}) {
	logkit.Errorf(format, args...)
}
