package http

import (
	"context"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/calendar/dto" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"        //nolint:depguard,nolintlint
)

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

// Application represents an interface of application visible to the server.
type Application interface {
	// CreateEvent is trying to build an Event object and save it in the storage.
	CreateEvent(ctx context.Context, input *dto.CreateEventInput) (*types.Event, error)

	// UpdateEvent is trying to get the existing Event from the storage, update it and save back.
	UpdateEvent(ctx context.Context, input *dto.UpdateEventInput) (*types.Event, error)

	// DeleteEvent is trying to delete the Event with the given ID from the storage.
	DeleteEvent(ctx context.Context, id string) error

	// GetEvent is trying to get the Event with the given ID from the storage.
	GetEvent(ctx context.Context, id string) (*types.Event, error)

	// GetAllUserEvents is trying to get all events for a given user ID from the storage.
	GetAllUserEvents(ctx context.Context, userID string) ([]*types.Event, error)

	// ListEvents is trying to get all events for a given user ID from the storage.
	ListEvents(ctx context.Context, input *dto.DateFilterInput) ([]*types.Event, error)

	// GetEventsForPeriod is trying to get all events for a given period from the storage.
	GetEventsForPeriod(ctx context.Context, input *dto.DateRangeInput) ([]*types.Event, error)
}
