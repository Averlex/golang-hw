package memory

import (
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types"                //nolint:depguard,nolintlint
)

type mutexMode int

// Lock types for helper calls.
const (
	readLock  mutexMode = iota // For using RLock and RUnlock respectively in read-only operations.
	writeLock                  // For using Lock and Unlock respectively in write-only operations.
)

// checkState checks if the storage is initialized and ready for operations.
// Returns ErrStorageUninitialized on failure.
func (s *Storage) checkState() error {
	if s.events == nil || s.userIndex == nil || s.idIndex == nil {
		return projectErrors.ErrStorageUninitialized
	}
	return nil
}

// deepCopySliceEvents creates a deep copy of the slice of events, element wise.
func deepCopySliceEvents(events []*types.Event) []*types.Event {
	res := make([]*types.Event, len(events))
	for i, event := range events {
		res[i] = types.DeepCopyEvent(event)
	}
	return res
}
