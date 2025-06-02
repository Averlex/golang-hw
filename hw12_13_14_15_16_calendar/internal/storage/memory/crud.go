package memory

import (
	"context"
	"fmt"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"                //nolint:depguard,nolintlint
	//nolint:depguard,nolintlint
)

// CreateEvent adds a new event to the in-memory storage. Method is imitation transactional behaviour,
// checking the context before applying changes.
//
// If the event already exists or the storage is full, it returns ErrDataExists or ErrStorageFull respectively.
//
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

	// Processing user events first due to potentially smaller number of events.
	if s.userIndex[event.UserID] == nil {
		s.userIndex[event.UserID] = []*types.Event{}
	}
	userPosition := s.findInsertPosition(s.userIndex[event.UserID], event)
	if s.isOverlaps(s.userIndex[event.UserID], event, userPosition) {
		err = projectErrors.ErrDateBusy
		return
	}
	position := s.findInsertPosition(s.events, event)

	// Context check before doing anything.
	if ctxErr := ctx.Err(); ctxErr != nil {
		err = fmt.Errorf("%w: %w", projectErrors.ErrTimeoutExceeded, ctxErr)
		return
	}

	// Applying changes.
	s.idIndex[event.ID] = event
	s.events = s.insertElem(s.events, event, position)
	s.userIndex[event.UserID] = s.insertElem(s.userIndex[event.UserID], event, userPosition)

	res = event
	return
}
