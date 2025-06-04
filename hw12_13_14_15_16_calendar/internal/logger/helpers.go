package logger

import (
	"log/slog"
	"reflect"
	"strings"
	"time"
)

// validationError is an error type for validation errors which follows error accumulation pattern.
type validationError struct {
	invalidTypes  []string
	invalidValues []string
}

func (e *validationError) Error() string {
	var b strings.Builder
	if len(e.invalidTypes) > 0 {
		b.WriteString("invalid_type=" + strings.Join(e.invalidTypes, ","))
	}
	if len(e.invalidValues) > 0 {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString("invalid_value=" + strings.Join(e.invalidValues, ","))
	}
	return b.String()
}

func (e *validationError) HasErrors() bool {
	return len(e.invalidTypes) > 0 || len(e.invalidValues) > 0
}

// validateLogLevel is a helper that checks if log level is valid.
func validateLogLevel(cfg map[string]any, ve *validationError) {
	if val, ok := cfg["level"]; ok {
		levelStr, ok := val.(string)
		if !ok {
			ve.invalidTypes = append(ve.invalidTypes, "level")
			return
		}
		levelStr = strings.ToLower(levelStr)

		switch levelStr {
		case "debug", "info", "warn", "error", "":
		default:
			ve.invalidValues = append(ve.invalidValues, "level")
		}
	}
}

// validateTimeFormat is a helper that checks if time format is valid.
func validateTimeFormat(cfg map[string]any, ve *validationError) {
	if val, ok := cfg["time_template"]; ok {
		timeTmpl, ok := val.(string)
		if !ok {
			ve.invalidTypes = append(ve.invalidTypes, "time_template")
			return
		}

		if timeTmpl == "" {
			return
		}

		testTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
		formatted := testTime.Format(timeTmpl)
		parsedTime, err := time.Parse(timeTmpl, formatted)
		if err != nil || !parsedTime.Equal(testTime) {
			ve.invalidValues = append(ve.invalidValues, "time_template")
		}
	}
}

// validateWriter is a helper that checks if writer is valid.
func validateWriter(cfg map[string]any, ve *validationError) {
	if val, ok := cfg["log_stream"]; ok {
		writerStr, ok := val.(string)
		if !ok {
			ve.invalidTypes = append(ve.invalidTypes, "log_stream")
			return
		}

		switch writerStr {
		case "stdout", "stderr", "":
		default:
			ve.invalidValues = append(ve.invalidValues, "log_stream")
		}
	}
}

// validateLogType is a helper that checks if log type is valid.
func validateLogType(cfg map[string]any, ve *validationError) {
	if val, ok := cfg["format"]; ok {
		logTypeStr, ok := val.(string)
		if !ok {
			ve.invalidTypes = append(ve.invalidTypes, "format")
			return
		}

		switch logTypeStr {
		case "json", "text", "":
		default:
			ve.invalidValues = append(ve.invalidValues, "format")
		}
	}
}

// validateFields returns missing and wrong type fields found in args.
// optionalFields is a map of field names with their expected types.
func validateTypes(args map[string]any, optionalFields map[string]any) (invalidTypes []string) {
	for field, expectedVal := range optionalFields {
		val, exists := args[field]
		if !exists {
			continue
		}

		// Default type switch will end up with false positive results. E.g., 123.(string) -> ok.
		// Using soft type check for string types, as, i.e., timeFormat in time package is untyped string.
		expectedKind := reflect.TypeOf(expectedVal).Kind()
		actualKind := reflect.TypeOf(val).Kind()
		if expectedKind != actualKind {
			invalidTypes = append(invalidTypes, field)
		}
	}

	return invalidTypes
}

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
