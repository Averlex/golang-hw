// Package logger package provides a constructor and wrapper methods
// for an underlying logger (currently - slog.Logger).
package logger

import (
	"errors"
	"io"
	"log/slog"
	"strings"
)

// Logger is a wrapper structure for an underlying logger.
type Logger struct {
	l *slog.Logger
}

var (
	// ErrInvalidLogType  is an error that is returned when the log type is invalid.
	ErrInvalidLogType = errors.New("invalid log type")
	// ErrInvalidLogLevel is an error that is returned when the log level is invalid.
	ErrInvalidLogLevel = errors.New("invalid log level")
	// ErrInvalidWriter is an error that is returned when the writer is not set.
	ErrInvalidWriter = errors.New("invalid writer set")
)

// New returns a new Logger with the given log type and level.
//
// The log type can be "text" or "json". The log level can be "debug", "info", "warn" or "error".
//
// Empty log level corresponds to "error", as well as empty log type corresponds to "json".
//
// If the log type or level is unknown, it returns an error.
func NewLogger(logType, level string, w io.Writer) (*Logger, error) {
	if w == nil {
		return nil, ErrInvalidWriter
	}

	logType = strings.ToLower(logType)
	level = strings.ToLower(level)

	var logHandler slog.Handler
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error", "":
		logLevel = slog.LevelError
	default:
		return nil, ErrInvalidLogLevel
	}

	switch logType {
	case "json", "":
		logHandler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: logLevel})
	case "text":
		logHandler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: logLevel})
	default:
		return nil, ErrInvalidLogType
	}
	return &Logger{slog.New(logHandler)}, nil
}

// Info logs a message with level Info on the standard logger.
func (logg Logger) Info(msg string, args ...any) {
	logg.l.Info(msg, args...)
}

// Error logs a message with level Error on the standard logger.
func (logg Logger) Error(msg string, args ...any) {
	logg.l.Error(msg, args...)
}

// Debug logs a message with level Debug on the standard logger.
func (logg Logger) Debug(msg string, args ...any) {
	logg.l.Debug(msg, args...)
}

// Warn logs a message with level Warn on the standard logger.
func (logg Logger) Warn(msg string, args ...any) {
	logg.l.Warn(msg, args...)
}

// With returns a new Logger that adds the given key-value pairs to the logger's context.
func (logg Logger) With(args ...any) *Logger {
	return &Logger{logg.l.With(args...)}
}
