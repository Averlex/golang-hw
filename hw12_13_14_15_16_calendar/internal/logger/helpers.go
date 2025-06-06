package logger

import (
	"log/slog"
	"strings"
	"time"
)

// buildHandler returns a handler based on log type.
func buildHandler(c *Config) slog.Handler {
	c.handlerOpts = &slog.HandlerOptions{
		Level: c.level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(c.timeTemplate))
				}
			}
			return a
		},
	}

	switch strings.ToLower(c.logType) {
	case "json", "":
		return slog.NewJSONHandler(c.writer, c.handlerOpts)
	case "text":
		return slog.NewTextHandler(c.writer, c.handlerOpts)
	default:
		return slog.NewJSONHandler(c.writer, c.handlerOpts)
	}
}
