package config

import (
	"fmt"

	calendarConfig "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/config/calendar" //nolint:depguard,nolintlint
)

type ServiceConfig interface {
	// GetSubConfig returns the part of the config that corresponds to the key.
	GetSubConfig(key string) (map[string]any, error)
}

// NewServiceConfig creates a concrete config based on service name.
func NewServiceConfig(name string) (ServiceConfig, error) {
	switch name {
	case "calendar":
		return &calendarConfig.Config{}, nil
	default:
		return nil, fmt.Errorf("unsupported service: %s", name)
	}
}
