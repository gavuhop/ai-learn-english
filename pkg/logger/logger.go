package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"ai-learn-english/config"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init initializes the logger with custom configuration
func init() {
	log = logrus.New()

	log.SetOutput(os.Stdout)

	level := config.Cfg.LogLevel
	switch level {
	case config.Debug:
		log.SetLevel(logrus.DebugLevel)
	case config.Info:
		log.SetLevel(logrus.InfoLevel)
	case config.Warn:
		log.SetLevel(logrus.WarnLevel)
	case config.Error:
		log.SetLevel(logrus.ErrorLevel)
	case config.Fatal:
		log.SetLevel(logrus.FatalLevel)
	case config.Panic:
		log.SetLevel(logrus.PanicLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
		ForceColors:     true,
		DisableColors:   false,
		DisableQuote:    true,
		DisableSorting:  false,
		PadLevelText:    true,
	})
}

// getCallerInfo returns the file and line number of the calling function
func getCallerInfo() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", 0
	}

	parts := strings.Split(file, "/")
	filename := parts[len(parts)-1]

	return filename, line
}

func Debug(format string, args ...interface{}) {
	file, line := getCallerInfo()
	log.WithFields(logrus.Fields{
		"file": fmt.Sprintf("%s:%d", file, line),
	}).Debugf(format, args...)
}

func Info(format string, args ...interface{}) {
	file, line := getCallerInfo()
	log.WithFields(logrus.Fields{
		"file": fmt.Sprintf("%s:%d", file, line),
	}).Infof(format, args...)
}

func Warn(format string, args ...interface{}) {
	file, line := getCallerInfo()
	log.WithFields(logrus.Fields{
		"file": fmt.Sprintf("%s:%d", file, line),
	}).Warnf(format, args...)
}

func Error(err error, format string, args ...interface{}) {
	file, line := getCallerInfo()

	fields := logrus.Fields{
		"file": fmt.Sprintf("%s:%d", file, line),
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	log.WithFields(fields).Errorf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	file, line := getCallerInfo()
	log.WithFields(logrus.Fields{
		"file": fmt.Sprintf("%s:%d", file, line),
	}).Errorf(format, args...)
}

func Fatal(err error, format string, args ...interface{}) {
	file, line := getCallerInfo()

	fields := logrus.Fields{
		"file": fmt.Sprintf("%s:%d", file, line),
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	log.WithFields(fields).Fatalf(format, args...)
}

// Fatalf logs a fatal message without error object and exits
func Fatalf(format string, args ...interface{}) {
	file, line := getCallerInfo()
	log.WithFields(logrus.Fields{
		"file": fmt.Sprintf("%s:%d", file, line),
	}).Fatalf(format, args...)
}

// WithField adds a field to the logger
func WithField(key string, value interface{}) *logrus.Entry {
	return log.WithField(key, value)
}

// WithFields adds multiple fields to the logger
func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}

// SetLevel sets the log level directly
func SetLevel(levelStr string) error {

	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}

	log.SetLevel(level)
	return nil
}

// GetLogger returns the underlying logrus logger
func GetLogger() *logrus.Logger {
	return log
}
