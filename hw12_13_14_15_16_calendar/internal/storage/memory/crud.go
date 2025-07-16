package memory

import (
	"context"
	"fmt"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                               //nolint:depguard,nolintlint
)

// CreateEvent adds a new event to the in-memory storage. Method is imitation transactional behaviour,
// checking the context before applying changes.
//
// If the event already exists or the storage is full, it returns ErrDataExists or ErrStorageFull respectively.
// If the event overlaps with another event, it returns ErrDateBusy.
//
// The event is inserted in a sorted order by Datetime, and if Datetime is equal,
// it uses ID for deterministic ordering.
func (s *Storage) CreateEvent(ctx context.Context, event *types.Event) (*types.Event, error) {
	method := "create event: %w"
	if event == nil {
		return nil, fmt.Errorf(method, projectErrors.ErrNoData)
	}

	var position, userPosition int // Positions for inserting the event in the inner data structure.

	err := s.withLockAndChecks(ctx, func() error {
		// Event with given ID already exists.
		if _, ok := s.idIndex[event.ID]; ok {
			return projectErrors.ErrDataExists
		}

		// Storage is already full.
		if len(s.events) == s.size {
			return projectErrors.ErrStorageFull
		}

		// Processing user events first due to potentially smaller number of events.
		if s.userIndex[event.UserID] == nil {
			s.userIndex[event.UserID] = []*types.Event{}
		}
		userPosition = s.findInsertPosition(s.userIndex[event.UserID], event)
		if s.isOverlaps(s.userIndex[event.UserID], event, userPosition) {
			return projectErrors.ErrDateBusy
		}
		position = s.findInsertPosition(s.events, event)
		return nil
	},
		func() {
			// Applying changes.
			s.idIndex[event.ID] = event
			s.events = s.insertElem(s.events, event, position)
			s.userIndex[event.UserID] = s.insertElem(s.userIndex[event.UserID], event, userPosition)
		},
		nil, writeLock)
	if err != nil {
		return nil, fmt.Errorf(method, err)
	}

	return types.DeepCopyEvent(event), nil
}

// UpdateEvent updates the event with the given ID in the in-memory storage.
// Method is imitation transactional behaviour, checking the context before applying changes.
//
// If the event does not exist, it returns ErrEventNotFound. If it overlaps with another event, it returns ErrDateBusy.
func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, data *types.EventData) (*types.Event, error) {
	method := "update event: %w"
	if data == nil {
		return nil, fmt.Errorf(method, projectErrors.ErrNoData)
	}

	var event *types.Event    // Event to update.
	var tmpEvent *types.Event // Temporary event to hold the updated data.
	var sourceIndex int       // Index of the event in the userIndex slice.
	var userPosition int      // Position for inserting the updated event in the userIndex slice.
	var isRollbackNeeded bool // Flag to indicate if rollback is needed.

	err := s.withLockAndChecks(ctx, func() error {
		var ok bool
		// Event with given ID not exists.
		if event, ok = s.idIndex[id]; !ok {
			return projectErrors.ErrEventNotFound
		}

		// Attempting to modify another user's event.
		if event.UserID != data.UserID {
			return projectErrors.ErrPermissionDenied
		}

		var err error
		tmpEvent, err = types.UpdateEvent(event.ID, data)
		if err != nil {
			return fmt.Errorf("unexpected error occurred: %w", err)
		}

		// Deleting the old event. No other inner structures are modified since the last possible error.
		sourceIndex = s.getIndex(s.userIndex[event.UserID], event)
		s.userIndex[event.UserID] = s.deleteElem(s.userIndex[event.UserID], sourceIndex)

		isRollbackNeeded = true // Since data was deleted, rollback is needed if the new event cannot be added.

		// Determining if the new event overlaps with existing events.
		userPosition = s.findInsertPosition(s.userIndex[event.UserID], tmpEvent)
		if s.isOverlaps(s.userIndex[event.UserID], tmpEvent, userPosition) {
			return projectErrors.ErrDateBusy
		}

		return nil
	}, func() {
		// Deleting old event data.
		delete(s.idIndex, event.ID)
		oldIndex := s.getIndex(s.events, event)
		s.events = s.deleteElem(s.events, oldIndex)

		// Adding new event data.
		s.idIndex[tmpEvent.ID] = tmpEvent
		newIndex := s.findInsertPosition(s.events, tmpEvent)
		s.events = s.insertElem(s.events, tmpEvent, newIndex)
		s.userIndex[tmpEvent.UserID] = s.insertElem(s.userIndex[tmpEvent.UserID], tmpEvent, userPosition)
	}, func() {
		if !isRollbackNeeded {
			return
		}
		s.userIndex[event.UserID] = s.insertElem(s.userIndex[event.UserID], event, sourceIndex)
	}, writeLock)
	if err != nil {
		return nil, fmt.Errorf(method, err)
	}

	return types.DeepCopyEvent(tmpEvent), nil
}

// DeleteEvent deletes the event with the given ID from the in-memory storage.
// Method is imitation transactional behaviour, checking the context before applying changes.
//
// If the event does not exist, it returns ErrEventNotFound.
func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	method := "delete event: %w"

	var event *types.Event // Event to delete.

	err := s.withLockAndChecks(ctx, func() error {
		var ok bool
		// Event with given ID not exists.
		if event, ok = s.idIndex[id]; !ok {
			return projectErrors.ErrEventNotFound
		}
		return nil
	}, func() {
		// Deleting old event data.
		delete(s.idIndex, event.ID)
		s.events = s.deleteElem(s.events, s.getIndex(s.events, event))
		s.userIndex[event.UserID] = s.deleteElem(s.userIndex[event.UserID], s.getIndex(s.userIndex[event.UserID], event))
		// User cache clean up.
		if len(s.userIndex[event.UserID]) == 0 {
			delete(s.userIndex, event.UserID)
		}
	}, nil, writeLock)
	if err != nil {
		return fmt.Errorf(method, err)
	}

	return nil
}

// UpdateNotifiedEvents updates the events with the given IDs in the storage as notified ones.
//
// If some of the IDs are not found in the DB, they will be ignored.
//
// Returns the number of updated events and nil on success, 0 and any error otherwise.
func (s *Storage) UpdateNotifiedEvents(ctx context.Context, notifiedEvents []uuid.UUID) (int64, error) {
	method := "update notified events: %w"
	if len(notifiedEvents) == 0 {
		return 0, fmt.Errorf(method, projectErrors.ErrNoData)
	}

	var updatedCount int64

	err := s.withLockAndChecks(ctx,
		nil,
		func() {
			// Updating the events with the given IDs if they exist in the storage.
			for _, id := range notifiedEvents {
				if _, ok := s.idIndex[id]; ok {
					s.idIndex[id].IsNotified = true
					updatedCount++
				}
			}
		},
		nil,
		writeLock,
	)
	if err != nil {
		return 0, fmt.Errorf(method, err)
	}

	return updatedCount, nil
}

// DeleteOldEvents deletes all events older than the given date from the database.
// Returns the number of deleted events and nil on success, 0 and any error otherwise.
func (s *Storage) DeleteOldEvents(ctx context.Context, date time.Time) (int64, error) {
	method := "delete old events: %w"

	var events []*types.Event // Events to delete.
	var deletedCount int64    // Number of deleted events.

	err := s.withLockAndChecks(
		ctx,
		func() error {
			// Creating a temporary event for getting the right index of the event to delete.
			tmpEvent, _ := types.UpdateEvent(uuid.Max, &types.EventData{Datetime: date})
			// Getting events to delete. If insert position == 0 -> all events are newer than requested.
			for i := range s.findInsertPosition(s.events, tmpEvent) {
				events = append(events, s.events[i])
			}
			return nil
		},
		func() {
			// This part might be optimized but it is kept as-is for simplicity due to rare storage clean up.
			for _, event := range events {
				// Deleting old event data.
				delete(s.idIndex, event.ID)
				s.events = s.deleteElem(s.events, s.getIndex(s.events, event))
				s.userIndex[event.UserID] = s.deleteElem(s.userIndex[event.UserID], s.getIndex(s.userIndex[event.UserID], event))
				deletedCount++
				// User cache clean up.
				if len(s.userIndex[event.UserID]) == 0 {
					delete(s.userIndex, event.UserID)
				}
			}
		},
		nil,
		writeLock,
	)
	if err != nil {
		return 0, fmt.Errorf(method, err)
	}

	return deletedCount, nil
}
