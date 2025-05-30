// Package storage provides a storage interface and factory method for storage construction.
package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memorystorage"        //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sqlstorage"           //nolint:depguard,nolintlint
	sttypes "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/storagetypes" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                                       //nolint:depguard,nolintlint
)

// Storage represents a universal storage interface.
type Storage interface {
	CreateEvent(ctx context.Context, event *sttypes.Event) (*sttypes.Event, error)
	UpdateEvent(ctx context.Context, id uuid.UUID, data *sttypes.EventData) (*sttypes.Event, error)
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	GetEventsForDay(ctx context.Context, date time.Time) ([]*sttypes.Event, error)
	GetEventsForWeek(ctx context.Context, date time.Time) ([]*sttypes.Event, error)
	GetEventsForMonth(ctx context.Context, date time.Time) ([]*sttypes.Event, error)
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
