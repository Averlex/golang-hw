// Package memory provides an in-memory storage implementation.
package memory

import (
	"context"
	"fmt"
	"sync"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                               //nolint:depguard,nolintlint
)

const defaultStorageSize = 10000 // Default maximum number of events in memory storage.

// Storage represents an in-memory storage for events.
type Storage struct {
	mu        sync.RWMutex
	size      int                        // Maximum number of events allowed.
	events    []*types.Event             // Sorted slice of events (by Datetime).
	idIndex   map[uuid.UUID]*types.Event // Index for fast lookup by event ID.
	userIndex map[string][]*types.Event  // Index for fast lookup by user ID.
}

// NewStorage creates a new in-memory Storage instance with a maximum event limit.
// If maxEvents is 0 or negative, a default limit of 10,000 is used.
// No initialization of data structures is performed until Connect is called.
func NewStorage(maxEvents int) (*Storage, error) {
	if maxEvents <= 0 {
		maxEvents = defaultStorageSize
	}
	return &Storage{
		size: maxEvents,
	}, nil
}

// Connect initializes the in-memory storage by creating the event slice and indexes.
// It checks the context before applying initialization.
// If the context is canceled, no changes are applied.
func (s *Storage) Connect(ctx context.Context) error {
	events := make([]*types.Event, 0)
	idIndex := make(map[uuid.UUID]*types.Event)
	userIndex := make(map[string][]*types.Event)

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("storage connection: %w: %w", projectErrors.ErrTimeoutExceeded, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = events
	s.idIndex = idIndex
	s.userIndex = userIndex
	return nil
}

// Close clears the in-memory storage and releases resources.
// It is safe to call multiple times.
func (s *Storage) Close(_ context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = nil
	s.idIndex = nil
	s.userIndex = nil
}
