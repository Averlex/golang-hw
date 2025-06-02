package memory

import (
	"context"
	"fmt"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"  //nolint:depguard,nolintlint
	"github.com/google/uuid"                                            //nolint:depguard,nolintlint
)

// GetEvent retrieves the event with the given ID from the in-memory storage.
// Method imitates transactional behavior, checking the context before returning the result.
//
// If the event does not exist, it returns nil and ErrEventNotFound.
func (s *Storage) GetEvent(ctx context.Context, id uuid.UUID) (*types.Event, error) {
	method := "get event: %w"

	var event *types.Event

	err := s.withLockAndChecks(ctx, func() error {
		// Event with given ID does not exist.
		var ok bool
		if event, ok = s.idIndex[id]; !ok {
			return errors.ErrEventNotFound
		}
		return nil
	}, nil, nil, readLock)
	if err != nil {
		return nil, fmt.Errorf(method, err)
	}

	return event, nil
}

// GetAllUserEvents retrieves all events for the given user from the in-memory storage.
// Method imitates transactional behavior, checking the context before returning the result.
//
// Returns a slice of events sorted by Datetime. If the user has no events, it returns an nil and ErrEventNotFound.
func (s *Storage) GetAllUserEvents(ctx context.Context, userID string) ([]*types.Event, error) {
	method := "get all user events: %w"

	var events []*types.Event

	err := s.withLockAndChecks(ctx, func() error {
		// No events for the user.
		var ok bool
		if events, ok = s.userIndex[userID]; !ok {
			return errors.ErrEventNotFound
		}
		return nil
	}, nil, nil, readLock)
	if err != nil {
		return nil, fmt.Errorf(method, err)
	}

	return events, nil
}
