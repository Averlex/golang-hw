package app

import (
	"context"
	"fmt"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                          //nolint:depguard,nolintlint
)

// Storage represents a universal storage interface.
type Storage interface {
	// Connect establishes a connection to the storage backend.
	Connect(ctx context.Context) error

	// CreateEvent creates a new event in the storage.
	// Returns the created event or an error if the operation fails.
	CreateEvent(ctx context.Context, event *types.Event) (*types.Event, error)

	// UpdateEvent updates an existing event by ID with the provided data.
	// Returns the updated event or an error if the operation fails.
	UpdateEvent(ctx context.Context, id uuid.UUID, data *types.EventData) (*types.Event, error)

	// DeleteEvent deletes an event by ID.
	// Returns an error if the operation fails.
	DeleteEvent(ctx context.Context, id uuid.UUID) error

	// GetEvent retrieves an event by ID.
	// Returns the event or an error if not found or the operation fails.
	GetEvent(ctx context.Context, id uuid.UUID) (*types.Event, error)

	// GetAllUserEvents retrieves all events for a given user ID.
	// Returns a slice of events or an error if not found or the operation fails.
	GetAllUserEvents(ctx context.Context, userID string) ([]*types.Event, error)

	// GetEventsForDay retrieves events for a specific day, optionally filtered by user ID.
	// Returns a slice of events or an error if not found or the operation fails.
	GetEventsForDay(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error)

	// GetEventsForWeek retrieves events for a specific week, optionally filtered by user ID.
	// Returns a slice of events or an error if not found or the operation fails.
	GetEventsForWeek(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error)

	// GetEventsForMonth retrieves events for a specific month, optionally filtered by user ID.
	// Returns a slice of events or an error if not found or the operation fails.
	GetEventsForMonth(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error)

	// GetEventsForPeriod retrieves events for a given period, optionally filtered by user ID.
	// Returns a slice of events or an error if not found or the operation fails.
	GetEventsForPeriod(ctx context.Context, dateStart, dateEnd time.Time, userID *string) ([]*types.Event, error)
}

// Logger represents an interface of logger visible to the app.
type Logger interface {
	// Info logs a message with level Info on the standard logger.
	Info(ctx context.Context, msg string, args ...any)
	// Debug logs a message with level Debug on the standard logger.
	Debug(ctx context.Context, msg string, args ...any)
	// Warn logs a message with level Warn on the standard logger.
	Warn(ctx context.Context, msg string, args ...any)
	// Error logs a message with level Error on the standard logger.
	Error(ctx context.Context, msg string, args ...any)
}

// CreateEvent is trying to build an Event object and save it in the storage.
// Returns *Event, nil on success, nil and error otherwise.
func (a *App) CreateEvent(ctx context.Context, input *types.CreateEventInput) (*types.Event, error) {
	msg := "create event: %w"
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
	err = a.withRetries(ctx, func() error {
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
func (a *App) UpdateEvent(ctx context.Context, input *types.UpdateEventInput) (*types.Event, error) {
	msg := "update event: %w"
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
	err = a.withRetries(ctx, func() error {
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
func (a *App) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	msg := "delete event: %w"

	// Trying to update the object in the storage.
	err := a.withRetries(ctx, func() error {
		err := a.s.DeleteEvent(ctx, id)
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
