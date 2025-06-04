// Package app provides calendar service with business logic handling.
package app

import (
	"fmt"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
)

// App represents a calendar application.
type App struct {
	l       Logger
	s       Storage
	retries int
}

// NewApp creates a new calendar application after arguments validation.
//
// It uses the provided logger and storage to log and store events.
func NewApp(logger Logger, storage Storage, config map[string]any) (*App, error) {
	// Args validation.
	missing := make([]string, 0)
	if logger == nil {
		missing = append(missing, "logger")
	}
	if storage == nil {
		missing = append(missing, "storage")
	}
	if config == nil {
		missing = append(missing, "config")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: some of the required parameters are missing: args=%v",
			errors.ErrAppInitFailed, missing)
	}

	// Field types validation.
	expectedFields := map[string]any{
		"retries": int(0),
	}
	missing, wrongType := validateFields(config, expectedFields)
	if len(missing) > 0 || len(wrongType) > 0 {
		return nil, fmt.Errorf("%w: missing=%v invalid_type=%v",
			errors.ErrCorruptedConfig, missing, wrongType)
	}

	// Extract from config an normalize the value.
	retries, _ := config["retries"].(int)
	retries = max(0, retries)

	return &App{
		l:       logger,
		s:       storage,
		retries: retries,
	}, nil
}
