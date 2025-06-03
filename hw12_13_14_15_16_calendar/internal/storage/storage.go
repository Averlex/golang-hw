// Package storage provides a storage interface and factory method for storage construction.
package storage

import (
	"fmt"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
)

// NewStorage creates a new storage instance based on the provided configuration.
// The args map must contain a "type" key specifying the storage type ("memory" or "sql").
// Returns an error wrapped with ErrCorruptedConfig if configuration is invalid,
// or ErrStorageInitFailed if initialization fails.
func NewStorage(args map[string]any) (Storage, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("%w: no storage configuration received", errors.ErrCorruptedConfig)
	}

	storageType, ok := args["type"]
	if !ok {
		return nil, fmt.Errorf("%w: no storage type received", errors.ErrCorruptedConfig)
	}

	var s Storage
	var err error

	switch storageType {
	case "memory":
		s, err = createMemoryStorage(args)
	case "sql":
		s, err = createSQLStorage(args)
	default:
		return nil, fmt.Errorf("%w: unknown storage type %q", errors.ErrCorruptedConfig, storageType)
	}

	if err != nil {
		return nil, fmt.Errorf("%q storage: %w", storageType, errors.ErrStorageInitFailed)
	}

	return s, nil
}
