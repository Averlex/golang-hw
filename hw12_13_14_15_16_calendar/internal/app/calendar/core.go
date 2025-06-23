package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/calendar/dto"         //nolint:depguard,nolintlint
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"                //nolint:depguard,nolintlint
)

// CreateEvent is trying to build an Event object and save it in the storage.
// Returns *Event, nil on success, nil and error otherwise.
func (a *App) CreateEvent(ctx context.Context, input *dto.CreateEventInput) (*types.Event, error) {
	method := "CreateEvent"
	msg := method + ": %w"
	if input == nil {
		return nil, fmt.Errorf(msg, projectErrors.ErrNoData)
	}

	// Constructing the Event object and validating it.
	event, err := types.NewEvent(
		input.Title,
		input.Datetime,
		input.Duration,
		safeDereference(input.Description),
		input.UserID,
		safeDereference(input.RemindIn),
	)
	if err != nil {
		return nil, fmt.Errorf(msg, err)
	}

	var resEvent *types.Event

	// Trying to save the object in the storage.
	err = a.withRetries(ctx, method, func() error {
		event, err := a.s.CreateEvent(ctx, event)
		if err != nil {
			return err
		}
		resEvent = event
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(msg, err)
	}

	return resEvent, nil
}

// UpdateEvent is trying to get the existing Event from the storage, update it and save back.
// Returns *Event, nil on success, nil and error otherwise.
func (a *App) UpdateEvent(ctx context.Context, input *dto.UpdateEventInput) (*types.Event, error) {
	method := "UpdateEvent"
	msg := method + ": %w"
	if input == nil {
		return nil, fmt.Errorf(msg, projectErrors.ErrNoData)
	}

	// Constructing the Event object and validating it.
	eventData, err := types.NewEventData(
		safeDereference(input.Title),
		safeDereference(input.Datetime),
		safeDereference(input.Duration),
		safeDereference(input.Description),
		safeDereference(input.UserID),
		safeDereference(input.RemindIn),
	)
	if err != nil {
		return nil, fmt.Errorf(msg, err)
	}

	var resEvent *types.Event

	// Trying to update the object in the storage.
	err = a.withRetries(ctx, method, func() error {
		event, err := a.s.UpdateEvent(ctx, input.ID, eventData)
		if err != nil {
			return err
		}
		resEvent = event
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(msg, err)
	}

	return resEvent, nil
}

// DeleteEvent is trying to delete the Event with the given ID from the storage.
// Returns nil on success and error otherwise.
func (a *App) DeleteEvent(ctx context.Context, id string) error {
	method := "DeleteEvent"
	msg := method + ": %w"

	uuidID, err := idFromString(id)
	if err != nil {
		return fmt.Errorf(msg, err)
	}

	// Trying to update the object in the storage.
	err = a.withRetries(ctx, method, func() error {
		err := a.s.DeleteEvent(ctx, *uuidID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf(msg, err)
	}

	return nil
}

// GetEvent is trying to get the Event with the given ID from the storage.
// Returns nil on success and error otherwise.
func (a *App) GetEvent(ctx context.Context, id string) (*types.Event, error) {
	method := "GetEvent"
	msg := method + ": %w"

	var resEvent *types.Event

	uuidID, err := idFromString(id)
	if err != nil {
		return nil, fmt.Errorf(msg, err)
	}

	// Trying to save the object in the storage.
	err = a.withRetries(ctx, method, func() error {
		event, err := a.s.GetEvent(ctx, *uuidID)
		if err != nil {
			return err
		}
		resEvent = event
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(msg, err)
	}

	return resEvent, nil
}

// GetAllUserEvents is trying to get all events for a given user ID from the storage.
func (a *App) GetAllUserEvents(ctx context.Context, userID string) ([]*types.Event, error) {
	method := "GetAllUserEvents"
	msg := method + ": %w"

	var events []*types.Event

	// Trying to save the object in the storage.
	err := a.withRetries(ctx, method, func() error {
		res, err := a.s.GetAllUserEvents(ctx, userID)
		if err != nil {
			return err
		}
		events = res
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(msg, err)
	}

	return events, nil
}

// ListEvents is trying to get all events for a given user ID from the storage.
//
// period is the period of time to get events for, stratring from the given date.
// Accepted values are Day, Week and Month.
//
// Returns []*Event, nil on success and nil, error otherwise.
//
// NOTE: period is casted to the the start of the corresponding calendar period.
func (a *App) ListEvents(ctx context.Context, input *dto.DateFilterInput) ([]*types.Event, error) {
	method := "ListEvents"
	msg := method + ": %w"

	if input == nil {
		return nil, fmt.Errorf(msg, projectErrors.ErrNoData)
	}

	// period validation. Incorrect value might be cause by programmer only.
	switch input.Period {
	case dto.Day, dto.Week, dto.Month:
	default:
		a.l.Error(ctx, "unexpected parameter value",
			slog.String("method", method),
			slog.String("parameter", "period"),
			slog.String("value", input.Period.String()),
			slog.Any("err", projectErrors.ErrInconsistentState),
		)
		return nil, fmt.Errorf(msg, projectErrors.ErrInconsistentState)
	}

	var events []*types.Event

	// Trying to save the object in the storage.
	err := a.withRetries(ctx, method, func() error {
		var res []*types.Event
		var err error
		switch input.Period {
		case dto.Day:
			res, err = a.s.GetEventsForDay(ctx, input.Date, input.UserID)
		case dto.Week:
			res, err = a.s.GetEventsForWeek(ctx, input.Date, input.UserID)
		case dto.Month:
			res, err = a.s.GetEventsForMonth(ctx, input.Date, input.UserID)
		}

		if err != nil {
			return err
		}
		events = res
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(msg, err)
	}

	return events, nil
}

// GetEventsForPeriod is trying to get all events for a given period from the storage.
// Returns []*Event, nil on success and nil, error otherwise.
//
// NOTE: time borders are not casted unlike in ListEvents.
func (a *App) GetEventsForPeriod(ctx context.Context, input *dto.DateRangeInput) ([]*types.Event, error) {
	method := "GetEventsForPeriod"
	msg := method + ": %w"

	if input == nil {
		return nil, fmt.Errorf(msg, projectErrors.ErrNoData)
	}

	var events []*types.Event

	// Trying to save the object in the storage.
	err := a.withRetries(ctx, method, func() error {
		res, err := a.s.GetEventsForPeriod(ctx, input.DateStart, input.DateEnd, input.UserID)
		if err != nil {
			return err
		}
		events = res
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(msg, err)
	}

	return events, nil
}
