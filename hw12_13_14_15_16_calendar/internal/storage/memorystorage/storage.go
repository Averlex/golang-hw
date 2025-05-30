// Package memorystorage provides an in-memory storage implementation.
package memorystorage

import (
	"context"
	"sync"
	"time"

	sttypes "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/storagetypes" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                                       //nolint:depguard,nolintlint
)

// Storage represents an in-memory storage. The implementation is concurrent-safe.
type Storage struct {
	// TODO
	mu sync.RWMutex //nolint:unused
}

// NewStorage creates a new Storage instance based on the given InMemoryConf.
//
// If the arguments are empty, it returns an error.
func NewStorage(_ any) (*Storage, error) {
	return &Storage{}, nil
}

//nolint:revive
func (*Storage) Connect(ctx context.Context) error { return nil }

//nolint:revive
func (*Storage) Close(ctx context.Context) error { return nil }

//nolint:revive
func (*Storage) CreateEvent(ctx context.Context, event *sttypes.Event) (*sttypes.Event, error) {
	return nil, nil
}

//nolint:revive
func (*Storage) UpdateEvent(ctx context.Context, id uuid.UUID, data *sttypes.EventData) (*sttypes.Event, error) {
	return nil, nil
}

//nolint:revive
func (*Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error { return nil }

//nolint:revive
func (*Storage) GetEvent(ctx context.Context, id uuid.UUID) (*sttypes.Event, error) { return nil, nil }

//nolint:revive
func (*Storage) GetAllUserEvents(ctx context.Context, userID string) ([]*sttypes.Event, error) {
	return nil, nil
}

//nolint:revive
func (*Storage) GetEventsForDay(ctx context.Context, date time.Time, userID *string) ([]*sttypes.Event, error) {
	return nil, nil
}

//nolint:revive
func (*Storage) GetEventsForWeek(ctx context.Context, date time.Time, userID *string) ([]*sttypes.Event, error) {
	return nil, nil
}

//nolint:revive
func (*Storage) GetEventsForMonth(ctx context.Context, date time.Time, userID *string) ([]*sttypes.Event, error) {
	return nil, nil
}

//nolint:revive
func (*Storage) GetEventsForPeriod(ctx context.Context, dateStart, dateEnd time.Time,
	userID *string,
) ([]*sttypes.Event, error) {
	return nil, nil
}
