// Package logger package provides a constructor and wrapper methods
// for an underlying logger (currently - slog.Logger).
package logger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Option defines a function that allows to configure underlying logger on construction.
type Option func(c *Config) error

// Config defines an inner logger configuration.
type Config struct {
	handlerOpts  *slog.HandlerOptions
	logType      string
	handler      slog.Handler
	writer       io.Writer
	timeTemplate string
	level        slog.Level
	setupLevel   bool
}

// checkDefaults sets default values for the logger configurations, if they are empty.
func (c *Config) checkDefaults() {
	if c.logType == "" {
		c.logType = DefaultLogType
	}
	if c.writer == nil {
		c.writer = DefaultWriterValue
	}
	if c.timeTemplate == "" {
		c.timeTemplate = DefaultTimeTemplate
	}
	if c.setupLevel {
		c.level = DefaultLevelValue
	}
}

// WithConfig allows to apply custom configuration.
// Expected following config structure:
//
//	{
//			format        string, // "text" or "json"
//			level         string, // "debug", "info", "warn", "error"
//			time_template string, // any valid time format
//			log_stream:   string, // "stdout", "stderr"
//	}
func WithConfig(cfg map[string]any) Option {
	return func(c *Config) error {
		optionalFields := map[string]any{
			"format":        "",
			"level":         "",
			"time_template": "",
			"log_stream":    "",
		}

		ve := &validationError{}
		ve.invalidTypes = validateTypes(cfg, optionalFields)

		validateLogLevel(cfg, ve)
		validateTimeFormat(cfg, ve)
		validateWriter(cfg, ve)
		validateLogType(cfg, ve)

		if ve.HasErrors() {
			return fmt.Errorf("config data is invalid: %s", ve.Error())
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
			case "error":
				c.level = slog.LevelError
			default:
				c.setupLevel = true
			}
		}

		if timeTmpl, ok := cfg["time_template"]; ok {
			c.timeTemplate = timeTmpl.(string)
		}

		if writer, ok := cfg["log_stream"]; ok {
			switch strings.ToLower(writer.(string)) {
			case "stdout":
				c.writer = os.Stdout
			case "stderr":
				c.writer = os.Stderr
			}
		}

		if logType, ok := cfg["format"]; ok {
			c.logType = logType.(string)
		}

		c.checkDefaults()
		c.handler = buildHandler(c)

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
		c.handler = buildHandler(c)

		return nil
	}
}

// SetDefaults is a wrapper over WithConfig, passing empty config.
func SetDefaults() Option {
	return WithConfig(map[string]any{
		"format":        DefaultLogType,
		"level":         DefaultLevel,
		"time_template": DefaultTimeTemplate,
		"log_stream":    DefaultWriter,
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

	if err := SetDefaults()(cfg); err != nil {
		return nil, errors.New("default logger initialization failed")
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("logger initialization failed: %w", err)
		}
	}

	return &Logger{slog.New(cfg.handler)}, nil
}
