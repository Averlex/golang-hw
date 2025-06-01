package memory

import (
	"context"
	"fmt"
	"os/user"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"
)

// CreateEvent adds a new event to the in-memory storage.
// It checks the context for cancellation before applying changes.
// If the event already exists or the storage is full, it returns ErrDataExists or ErrStorageFull respectively.
// The event is inserted in a sorted order by Datetime, and if Datetime is equal,
// it uses ID for deterministic ordering.
func (s *Storage) CreateEvent(ctx context.Context, event *types.Event) (res *types.Event, err error) {
	// Local error wrapping helper.
	defer func() {
		if err != nil {
			res = nil
			err = fmt.Errorf("create event: %w", err)
		}
	}()

	if event == nil {
		err = projectErrors.ErrNoData
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Storage init check.
	err = s.checkState()
	if err != nil {
		return
	}

	// Event with given ID already exists.
	if _, ok := s.idIndex[event.ID]; ok {
		err = projectErrors.ErrDataExists
		return
	}

	// Storage is already full.
	if len(s.events) == s.size {
		err = projectErrors.ErrStorageFull
		return
	}

	// Preparing changes.
	position := s.findInsertPosition(s.events, event)
	//nolint:gocritic
	events := append(s.events[:position], append([]*types.Event{event}, s.events[position:]...)...)
	if s.userIndex[event.UserID] == nil {
		s.userIndex[event.UserID] = []*types.Event{}
	}
	userPosition := s.findInsertPosition(s.userIndex[event.UserID], event)
	//nolint:gocritic
	userEvents := append(s.userIndex[event.UserID][:userPosition],
		append([]*types.Event{event}, s.userIndex[event.UserID][userPosition:]...)...)

	// Context check before applying changes.
	if ctxErr := ctx.Err(); ctxErr != nil {
		err = fmt.Errorf("%w: %w", projectErrors.ErrTimeoutExceeded, ctxErr)
		return
	}

	// Applying changes.
	s.idIndex[event.ID] = event
	s.events = events
	s.userIndex[event.UserID] = userEvents

	res = event
	return
}

func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, data *types.EventData) (res *types.Event, err error) {
	// Local error wrapping helper.
	defer func() {
		if err != nil {
			res = nil
			err = fmt.Errorf("create event: %w", err)
		}
	}()

	if data == nil {
		err = projectErrors.ErrNoData
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Storage init check.
	err = s.checkState()
	if err != nil {
		return
	}

	var event *types.Event
	var ok bool
	// Event with given ID already exists.
	if event, ok = s.idIndex[id]; !ok {
		err = projectErrors.ErrEventNotFound
		return
	}

	res = event

	// Preparing changes.
	userEvents := s.userIndex[event.UserID]
	

	// Context check before applying changes.
	if ctxErr := ctx.Err(); ctxErr != nil {
		err = fmt.Errorf("%w: %w", projectErrors.ErrTimeoutExceeded, ctxErr)
		return
	}

	// Applying changes.
	

}
