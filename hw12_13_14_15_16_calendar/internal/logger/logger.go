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

// Option defines a function that allows to configure underlying logger on construction.
type Option func(c *Config) error

// Config defines an inner logger configuration.
type Config struct {
	handlerOpts  *slog.HandlerOptions
	handler      slog.Handler
	writer       io.Writer
	timeTemplate string
	level        slog.Level
}

func (c *Config) getConfig() map[string]any {
	return map[string]any{
		"logType":      c.handler,
		"level":        c.level,
		"timeTemplate": c.timeTemplate,
		"writer":       c.writer,
	}
}

// WithConfig allows to apply custom configuration.
func WithConfig(cfg map[string]any) Option {
	return func(c *Config) error {
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
			return fmt.Errorf("%w: %s", errors.ErrCorruptedConfig, ve.Error())
		}

		if level, ok := cfg["level"]; ok {
			levelStr := strings.ToLower(level.(string))
			switch levelStr {
			case "debug":
				c.level = slog.LevelDebug
			case "info":
				c.level = slog.LevelInfo
			case "warn":
				c.level = slog.LevelWarn
			case "error", "":
				c.level = slog.LevelError
			}
		}

		if timeTmpl, ok := cfg["timeTemplate"]; ok {
			c.timeTemplate = timeTmpl.(string)
		} else {
			c.timeTemplate = defaultTimeTemplate
		}

		// Building arg for handler constructor.
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
func WithWriter(w io.Writer) Option {
	return func(c *Config) error {
		if w == nil {
			return fmt.Errorf("expected io.Writer, got nil")
		}

		c.writer = w

		return WithConfig(c.getConfig())(c)
	}
}

// SetDefaults is a wrapper over WithConfig, passing empty config.
func SetDefaults() Option {
	return WithConfig(map[string]any{
		"logType":      "",
		"level":        "",
		"timeTemplate": "",
		"writer":       "",
	})
}

// NewLogger returns a new Logger with the given log type and level.
// If no opts are provided, it returns a default logger.
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
func NewLogger(opts ...Option) (*Logger, error) {
	cfg := &Config{}
	SetDefaults()(cfg)
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrLoggerInitFailed, err.Error())
		}
	}

	return &Logger{slog.New(cfg.handler)}, nil
}
