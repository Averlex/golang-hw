// Package memorystorage provides an in-memory storage implementation.
package memorystorage

import "sync"

// Storage represents an in-memory storage. The implementation is concurrent-safe.
type Storage struct {
	// TODO
	mu sync.RWMutex //nolint:unused
}

// NewStorage creates a new Storage instance based on the given InMemoryConf.
//
// If the arguments are empty, it returns an error.
func NewStorage() *Storage {
	return &Storage{}
}

// TODO
