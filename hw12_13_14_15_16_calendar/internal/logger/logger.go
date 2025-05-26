// Package logger package provides a constructor and wrapper methods
// for an underlying logger (currently - slog.Logger).
package logger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// Logger is a wrapper structure for an underlying logger.
type Logger struct {
	l *slog.Logger
}

const defaultTimeTemplate = "02.01.2006 15:04:05.000"

var (
	// ErrInvalidLogType  is an error that is returned when the log type is invalid.
	ErrInvalidLogType = errors.New("invalid log type")
	// ErrInvalidLogLevel is an error that is returned when the log level is invalid.
	ErrInvalidLogLevel = errors.New("invalid log level")
	// ErrInvalidWriter is an error that is returned when the writer is not set.
	ErrInvalidWriter = errors.New("invalid writer set")
	// ErrInvalidTimeTemplate is an error that is returned when the time template cannot be parsed by time package.
	ErrInvalidTimeTemplate = errors.New("invalid time template")
)

// NewLogger returns a new Logger with the given log type and level.
//
// The log type can be "text" or "json". The log level can be "debug", "info", "warn" or "error".
//
// timeTemplate is a time format string. Any format which is valid for time.Time format is acceptable.
//
// Empty log level corresponds to "error", as well as empty log type corresponds to "json".
// Empty time format is equal to the default value which is "02.01.2006 15:04:05.000".
//
// If the log type or level is unknown, it returns an error.
func NewLogger(logType, level, timeTemplate string, w io.Writer) (*Logger, error) {
	if w == nil {
		return nil, ErrInvalidWriter
	}

	logType = strings.ToLower(logType)
	level = strings.ToLower(level)

	var logHandler slog.Handler
	var logLevel slog.Level

	// Log level validation.
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
		return nil, fmt.Errorf("%w: %s", ErrInvalidLogLevel, level)
	}

	// Time format validation.
	switch timeTemplate {
	case "":
		timeTemplate = defaultTimeTemplate
	default:
		testTime := time.Now()
		if _, err := time.Parse(timeTemplate, testTime.Format(timeTemplate)); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidTimeTemplate, timeTemplate)
		}
	}

	// Setting up log handlers options.
	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(timeTemplate))
				}
			}
			return a
		},
	}

	// Log type validation.
	switch logType {
	case "json", "":
		logHandler = slog.NewJSONHandler(w, opts)
	case "text":
		logHandler = slog.NewTextHandler(w, opts)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidLogType, logType)
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
