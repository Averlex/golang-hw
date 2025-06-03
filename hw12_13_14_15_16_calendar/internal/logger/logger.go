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

const defaultTimeTemplate = "02.01.2006 15:04:05.000"

// LoggerOption defines a function that allows to configure underlying logger on construction.
type LoggerOption func(c *loggerConfig) error

// loggerConfig defines an inner logger configuration.
type loggerConfig struct {
	handlerOpts  *slog.HandlerOptions
	handler      slog.Handler
	writer       io.Writer
	logType      string
	timeTemplate string
	logLevel     slog.Level
}

// WithConfig allows to apply custom configuration.
func WithConfig(cfg map[string]any) LoggerOption {
	return func(c *loggerConfig) error {
		optionalFields := map[string]any{
			"logType":      "",
			"level":        "",
			"timeTemplate": "",
			"writer":       "",
		}

		ve := &validationError{}

		ve.invalidTypes = validateTypes(cfg, optionalFields)

		validateLogLevel(cfg, ve)
		validateTimeFormat(cfg, ve)
		validateWriter(cfg, ve)
		validateLogType(cfg, ve)

		if ve.HasErrors() {
			return fmt.Errorf("%w: %s", errors.ErrCorruptedConfig, ve)
		}

		if level, ok := cfg["level"]; ok {
			levelStr := strings.ToLower(level.(string))
			switch levelStr {
			case "debug":
				c.logLevel = slog.LevelDebug
			case "info":
				c.logLevel = slog.LevelInfo
			case "warn":
				c.logLevel = slog.LevelWarn
			case "error", "":
				c.logLevel = slog.LevelError
			}
		}

		if timeTmpl, ok := cfg["timeTemplate"]; ok {
			c.timeTemplate = timeTmpl.(string)
		} else {
			c.timeTemplate = defaultTimeTemplate
		}

		// Building arg for handler constructor.
		c.handlerOpts = &slog.HandlerOptions{
			Level: c.logLevel,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					if t, ok := a.Value.Any().(time.Time); ok {
						a.Value = slog.StringValue(t.Format(c.timeTemplate))
					}
				}
				return a
			},
		}

		if writer, ok := cfg["writer"]; ok {
			switch strings.ToLower(writer.(string)) {
			case "stdout", "":
				c.writer = os.Stdout
			case "stderr":
				c.writer = os.Stderr
			}
		} else {
			c.writer = os.Stdout
		}

		if logType, ok := cfg["logType"]; ok {
			switch strings.ToLower(logType.(string)) {
			case "json", "":
				c.handler = slog.NewJSONHandler(c.writer, c.handlerOpts)
			case "text":
				c.handler = slog.NewTextHandler(c.writer, c.handlerOpts)
			}
		} else {
			c.handler = slog.NewJSONHandler(c.writer, c.handlerOpts)
		}

		return nil
	}
}

// WithWriter allows to apply custom configuration.
func WithWriter(w io.Writer) LoggerOption {
	return func(c *loggerConfig) error {
		if w == nil {
			return fmt.Errorf("expected io.Writer, got nil")
		}

		c.writer = w

		return nil
	}
}

// SetDefaults is a wrapper over WithConfig, passing empty config.
func SetDefaults() LoggerOption {
	return WithConfig(map[string]any{
		"logType":      "",
		"level":        "",
		"timeTemplate": "",
		"writer":       "",
	})
}

// NewLogger returns a new Logger with the given log type and level. If no opts are provided, it returns a default logger.
//
// The log type can be "text" or "json". The log level can be "debug", "info", "warn" or "error".
//
// timeTemplate is a time format string. Any format which is valid for time.Time format is acceptable.
//
// Empty log level corresponds to "error", as well as empty log type corresponds to "json".
// Empty time format is equal to the default value which is "02.01.2006 15:04:05.000".
// Empty writer option equals to using os.Stdout. Custom writer might be set using WithWriter option.
//
// If the log type or level is unknown, it returns an error.
func NewLogger(opts ...LoggerOption) (*Logger, error) {
	cfg := &loggerConfig{}
	SetDefaults()(cfg)
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrLoggerInitFailed, err)
		}
	}

	return &Logger{slog.New(cfg.handler)}, nil
}
