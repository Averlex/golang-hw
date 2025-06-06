package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger is a wrapper structure for an underlying logger.
type Logger struct {
	l *slog.Logger
}

func (logg Logger) addRequestContext(ctx context.Context, args ...any) []any {
	if v := ctx.Value("request"); v != nil {
		if attr, ok := v.(slog.Attr); ok {
			args = append(args, attr)
		}
	}

	return args
}

// Info logs a message with level Info on the standard logger.
func (logg Logger) Info(ctx context.Context, msg string, args ...any) {
	logg.l.Info(msg, logg.addRequestContext(ctx, args...)...)
}

// Error logs a message with level Error on the standard logger.
func (logg Logger) Error(ctx context.Context, msg string, args ...any) {
	logg.l.Error(msg, logg.addRequestContext(ctx, args...)...)
}

// Debug logs a message with level Debug on the standard logger.
func (logg Logger) Debug(ctx context.Context, msg string, args ...any) {
	logg.l.Debug(msg, logg.addRequestContext(ctx, args...)...)
}

// Warn logs a message with level Warn on the standard logger.
func (logg Logger) Warn(ctx context.Context, msg string, args ...any) {
	logg.l.Warn(msg, logg.addRequestContext(ctx, args...)...)
}

// Fatal logs a message with level Error on the standard logger and then calls os.Exit(1).
func (logg Logger) Fatal(ctx context.Context, msg string, args ...any) {
	logg.l.Error(msg, logg.addRequestContext(ctx, args...)...)
	os.Exit(1)
}

// With returns a new Logger that adds the given key-value pairs to the logger's context.
func (logg Logger) With(args ...any) *Logger {
	return &Logger{logg.l.With(args...)}
}
