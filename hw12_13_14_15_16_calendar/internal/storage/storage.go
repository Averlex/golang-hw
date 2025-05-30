// Package storage provides a storage interface and factory method for storage construction.
package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	memorystorage "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory" //nolint:depguard,nolintlint
	sqlstorage "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"       //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                                       //nolint:depguard,nolintlint
)

// Storage represents a universal storage interface.
type Storage interface {
	// Connect establishes a connection to the storage backend.
	Connect(ctx context.Context) error

	// Close closes the connection to the storage backend.
	Close(ctx context.Context) error

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

//nolint:revive,funlen,nolintlint
func NewStorage(args map[string]any) (Storage, error) {
	if len(args) == 0 {
		return nil, errors.New("no storage configuration reveived")
	}

	storageType, ok := args["type"]
	if !ok {
		return nil, errors.New("no storage type reveived")
	}

	var s Storage
	var err error
	var timeout time.Duration
	var errArgs []string

	// Timeout validation.
	if v, ok := args["timeout"]; !ok {
		errArgs = append(errArgs, "timeout")
	} else {
		timeout, ok = v.(time.Duration)
		if !ok {
			errArgs = append(errArgs, "timeout")
		} else {
			// Ensure timeout is not negative.
			timeout = max(0, timeout)
		}
	}

	switch storageType {
	case "memory":
		s, err = memorystorage.NewStorage(timeout)

	case "sql":
		sqlArgs, ok := args["sql"].(map[string]any)
		if !ok {
			return nil, errors.New("no sql storage configuration reveived")
		}
		callArgs := map[string]string{
			"host":     "",
			"port":     "",
			"user":     "",
			"password": "",
			"dbname":   "",
			"driver":   "postgres",
		}
		for k, v := range sqlArgs {
			val, ok := v.(string)
			if !ok {
				errArgs = append(errArgs, k)
				continue
			}
			callArgs[k] = val
		}

		if len(errArgs) > 0 {
			return nil, fmt.Errorf("invalid sql storage arguments: %v", errArgs)
		}

		s, err = sqlstorage.NewStorage(timeout, callArgs["driver"], callArgs["host"], callArgs["port"],
			callArgs["user"], callArgs["password"], callArgs["dbname"])

	default:
		return nil, fmt.Errorf("unknown storage type %v", storageType)
	}

	if err != nil {
		return nil, err
	}

	return s, nil
}
