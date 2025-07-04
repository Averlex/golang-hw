// Package config provides configuration loader and the interface for working with configuration.
package config

import (
	"errors"
)

// Additional errors to control the execution flow.
var (
	// ErrShouldStop is returned when the execution should be stopped. E.g. on -v and -h flags.
	ErrShouldStop = errors.New("execution should be stopped")
)

// ServiceConfig is an interface generalizing service config.
type ServiceConfig interface {
	// GetSubConfig returns the part of the config that corresponds to the key.
	GetSubConfig(key string) (map[string]any, error)
}
