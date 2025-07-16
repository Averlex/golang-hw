package storage

import (
	"context"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                //nolint:depguard,nolintlint
)

// Storage represents a universal storage interface.
type Storage interface {
	// Connect establishes a connection to the storage backend.
	Connect(ctx context.Context) error

	// Close closes the connection to the storage backend.
	Close(ctx context.Context)

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

	// GetEventsForNotification retrieves events for notification, optionally filtered by user ID.
	// Returns a slice of events or an error if not found or the operation fails.
	GetEventsForNotification(ctx context.Context) ([]*types.Event, error)

	// UpdateNotifiedEvents updates notified events in the storage.
	// Returns the number of updated events or an error if the operation fails.
	UpdateNotifiedEvents(ctx context.Context, notifiedEvents []uuid.UUID) (int64, error)

	// DeleteOldEvents deletes old events from the storage.
	// Returns the number of deleted events or an error if the operation fails.
	DeleteOldEvents(ctx context.Context, date time.Time) (int64, error)
}
