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
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return replaceTimeAttrs(groups, a, c.timeTemplate)
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

func replaceTimeAttrs(groups []string, a slog.Attr, timeFormat string) slog.Attr {
	// Default log tilestamp.
	if a.Key == slog.TimeKey && len(groups) == 0 {
		if t, ok := a.Value.Any().(time.Time); ok {
			a.Value = slog.StringValue(t.Format(timeFormat))
		}
		return a
	}

	// Common time.Time fields.
	if v, ok := a.Value.Any().(time.Time); ok {
		a.Value = slog.StringValue(v.Format(timeFormat))
		return a
	}

	// Groups handling.
	if a.Value.Kind() == slog.KindGroup {
		newGroup := make([]slog.Attr, len(a.Value.Group()))
		for i, ga := range a.Value.Group() {
			newGroup[i] = replaceTimeAttrs(append(groups, a.Key), ga, timeFormat)
		}
		a.Value = slog.GroupValue(newGroup...)
	}

	return a
}
