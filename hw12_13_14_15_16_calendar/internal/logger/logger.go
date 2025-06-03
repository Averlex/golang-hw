// Package logger package provides a constructor and wrapper methods
// for an underlying logger (currently - slog.Logger).
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
)

// Logger is a wrapper structure for an underlying logger.
type Logger struct {
	l *slog.Logger
}

const defaultTimeTemplate = "02.01.2006 15:04:05.000"

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
		return nil, errors.ErrInvalidWriter
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
		return nil, fmt.Errorf("%w: %s", errors.ErrInvalidLogLevel, level)
	}

	// Time format validation.
	switch timeTemplate {
	case "":
		timeTemplate = defaultTimeTemplate
	default:
		// Trying to format the test time and parse it afterwards.
		testTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
		formatted := testTime.Format(timeTemplate)
		parsedTime, err := time.Parse(timeTemplate, formatted)
		if err != nil || !parsedTime.Equal(testTime) {
			return nil, fmt.Errorf("%w: %s", errors.ErrInvalidTimeTemplate, timeTemplate)
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
		return nil, fmt.Errorf("%w: %s", errors.ErrInvalidLogType, logType)
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

// Fatal logs a message with level Error on the standard logger and then calls os.Exit(1).
func (logg Logger) Fatal(msg string, args ...any) {
	logg.l.Error(msg, args...)
	os.Exit(1)
}

// With returns a new Logger that adds the given key-value pairs to the logger's context.
func (logg Logger) With(args ...any) *Logger {
	return &Logger{logg.l.With(args...)}
}
