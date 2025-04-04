package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// Global logger instance
var (
	Log *logrus.Logger
)

// Log levels
const (
	DebugLevel = logrus.DebugLevel
	InfoLevel  = logrus.InfoLevel
	WarnLevel  = logrus.WarnLevel
	ErrorLevel = logrus.ErrorLevel
	FatalLevel = logrus.FatalLevel
	PanicLevel = logrus.PanicLevel
)

// Fields type for structured logging
type Fields logrus.Fields

// Init initializes the logger with specified configuration
func Init(level logrus.Level, logFilePath string) {
	Log = logrus.New()
	Log.SetLevel(level)
	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// Create multi-writer to log to both file and stdout
	outputs := []io.Writer{os.Stdout}

	// Add file output if logFilePath is provided
	if logFilePath != "" {
		err := os.MkdirAll(filepath.Dir(logFilePath), 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		} else {
			file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			} else {
				outputs = append(outputs, file)
			}
		}
	}

	// Set output to multi-writer
	Log.SetOutput(io.MultiWriter(outputs...))
}

// WithFields returns a logger entry with fields
func WithFields(fields Fields) *logrus.Entry {
	return Log.WithFields(logrus.Fields(fields))
}

// WithField returns a logger entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Fatalf(format, args...)
}

// Panic logs a panic message and panics
func Panic(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Panic(args...)
}

// Panicf logs a formatted panic message and panics
func Panicf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Panicf(format, args...)
}
