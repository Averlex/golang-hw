package logger

import (
	"log/slog"
	"os"
)

// Logger is a wrapper structure for an underlying logger.
type Logger struct {
	l *slog.Logger
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
