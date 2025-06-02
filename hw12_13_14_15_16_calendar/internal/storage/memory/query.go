package memory

import (
	"context"
	"fmt"
	"time"

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

// GetEventsForPeriod retrieves events within the specified time period from the in-memory storage.
// If userID is provided, it filters events for that user; otherwise, it returns events for all users.
// Method imitates transactional behavior, checking the context before returning the result.
//
// Returns a slice of events sorted by Datetime, considering only events that start within [dateStart, dateEnd].
// If no events are found, it returns nil and ErrEventNotFound.
func (s *Storage) GetEventsForPeriod(ctx context.Context,
	dateStart, dateEnd time.Time,
	userID *string,
) ([]*types.Event, error) {
	method := "get events for period: %w"

	var events []*types.Event

	err := s.withLockAndChecks(ctx, func() error {
		var sourceEvents []*types.Event
		if userID != nil {
			if userEvents, ok := s.userIndex[*userID]; ok {
				sourceEvents = userEvents
			}
		} else {
			sourceEvents = s.events
		}

		if len(sourceEvents) == 0 {
			return errors.ErrEventNotFound
		}

		// Create temporary events for binary search.
		startEvent, _ := types.UpdateEvent(uuid.New(), &types.EventData{Datetime: dateStart})
		endEvent, _ := types.UpdateEvent(uuid.New(), &types.EventData{Datetime: dateEnd})

		// Find left boundary: first event where Datetime >= dateStart.
		leftIdx := s.findInsertPosition(sourceEvents, startEvent)

		// Find right boundary: first event where Datetime > dateEnd.
		rightIdx := s.findInsertPosition(sourceEvents, endEvent)

		if leftIdx >= len(sourceEvents) || leftIdx >= rightIdx {
			return errors.ErrEventNotFound
		}

		events = sourceEvents[leftIdx:rightIdx]
		return nil
	}, nil, nil, readLock)
	if err != nil {
		return nil, fmt.Errorf(method, err)
	}

	return events, nil
}

// GetEventsForDay retrieves events for the specified day from the in-memory storage.
// If userID is provided, it filters events for that user; otherwise, it returns events for all users.
// Method imitates transactional behavior, checking the context before returning the result.
//
// Returns a slice of events sorted by Datetime. If no events are found, it returns nil and ErrEventNotFound.
func (s *Storage) GetEventsForDay(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error) {
	dateStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dateEnd := dateStart.AddDate(0, 0, 1)

	res, err := s.GetEventsForPeriod(ctx, dateStart, dateEnd, userID)
	if err != nil {
		return nil, fmt.Errorf("get events for day: %w", err)
	}

	return res, nil
}

// GetEventsForWeek retrieves events for the week containing the specified date from the in-memory storage.
// If userID is provided, it filters events for that user; otherwise, it returns events for all users.
// Method imitates transactional behavior, checking the context before returning the result.
//
// Returns a slice of events sorted by Datetime. If no events are found, it returns nil and ErrEventNotFound.
func (s *Storage) GetEventsForWeek(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error) {
	// Weekday considering Monday as the first day of the week.
	weekday := (int(date.Weekday()-time.Monday) + 7) % 7

	// Truncating the date to the start of the week.
	dateStart := date.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
	dateEnd := dateStart.AddDate(0, 0, 7)

	res, err := s.GetEventsForPeriod(ctx, dateStart, dateEnd, userID)
	if err != nil {
		return nil, fmt.Errorf("get events for week: %w", err)
	}

	return res, nil
}

// GetEventsForMonth retrieves events for the month containing the specified date from the in-memory storage.
// If userID is provided, it filters events for that user; otherwise, it returns events for all users.
// Method imitates transactional behavior, checking the context before returning the result.
//
// Returns a slice of events sorted by Datetime. If no events are found, it returns nil and ErrEventNotFound.
func (s *Storage) GetEventsForMonth(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error) {
	// Truncating the date to the start of the month.
	dateStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	dateEnd := dateStart.AddDate(0, 1, 0)

	res, err := s.GetEventsForPeriod(ctx, dateStart, dateEnd, userID)
	if err != nil {
		return nil, fmt.Errorf("get events for month: %w", err)
	}

	return res, nil
}
