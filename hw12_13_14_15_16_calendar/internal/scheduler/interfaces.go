package scheduler

import (
	"context"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                //nolint:depguard,nolintlint
)

// Storage represents a universal storage interface.
type Storage interface {
	// Connect establishes a connection to the storage backend.
	Connect(context.Context) error

	// GetEventsForNotification retrieves events for notification, optionally filtered by user ID.
	// Returns a slice of events or an error if not found or the operation fails.
	GetEventsForNotification(context.Context) ([]*types.Event, error)

	// UpdateNotifiedEvents updates notified events in the storage.
	// Returns the number of updated events or an error if the operation fails.
	UpdateNotifiedEvents(context.Context, []uuid.UUID) (int64, error)

	// DeleteOldEvents deletes old events from the storage.
	// Returns the number of deleted events or an error if the operation fails.
	DeleteOldEvents(context.Context, time.Time) (int64, error)
}

// Logger represents an interface of logger visible to the app.
type Logger interface {
	// Info logs a message with level Info on the standard logger.
	Info(context.Context, string, ...any)
	// Debug logs a message with level Debug on the standard logger.
	Debug(context.Context, string, ...any)
	// Warn logs a message with level Warn on the standard logger.
	Warn(context.Context, string, ...any)
	// Error logs a message with level Error on the standard logger.
	Error(context.Context, string, ...any)
}

// MessageBroker represents a universal message broker interface.
type MessageBroker interface {
	// Produce sends a message to the message broker.
	// Returns an error if the operation fails.
	Produce(context.Context, []byte) error
}
