// Package sql provides a SQL database storage implementation.
package sql

import (
	"context"
	"fmt"
	"sync"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	_ "github.com/lib/pq"                                                             //nolint:depguard,nolintlint
)

// Storage represents a SQL database storage.
type Storage struct {
	mu      sync.RWMutex
	driver  string
	db      DB
	dsn     string
	timeout time.Duration
}

// StorageOption defines a function that allows to configure underlying Storage DB on construction.
// Use it for testing purposes to inject a custom DB implementation.
type StorageOption func(s *Storage)

// WithDB allows to inject a custom DB implementation (for testing).
func WithDB(db DB) StorageOption {
	return func(s *Storage) {
		s.db = db
	}
}

// NewStorage creates a new Storage instance based on the given args.
//
// If the arguments are empty, it returns an error.
//
// The function constructs a DSN based on the given arguments and
// the default driver. No connection is established upon the call.
//
// Currently supported drivers are "postgres" or "postgresql".
func NewStorage(timeout time.Duration, driver, host, port, user, password, dbname string,
	opts ...StorageOption,
) (*Storage, error) {
	//nolint:goconst,nolintlint
	if driver != "postgres" && driver != "postgresql" {
		return nil, projectErrors.ErrUnsupportedDriver
	}

	// Normalizing driver name to "postgres" for consistency.
	driver = "postgres"

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d",
		host, port, user, password, dbname, int(timeout.Seconds()),
	)

	storage := &Storage{
		db:      &SQLXWrapper{},
		driver:  driver,
		dsn:     dsn,
		timeout: timeout,
	}

	for _, opt := range opts {
		opt(storage)
	}
	return storage, nil
}

// Connect connects to the database.
//
// If the connection is successful, it pings the database
// to check if the connection is alive. If any error occurs during the connection
// or pinging, it returns an error.
func (s *Storage) Connect(ctx context.Context) error {
	return s.withTimeout(ctx, func(localCtx context.Context) error {
		_, err := s.db.ConnectContext(localCtx, s.driver, s.dsn)
		if err != nil {
			return fmt.Errorf("storage connection: %w", err)
		}
		return nil
	})
}

// Close closes the connection to the database.
// Method is safe to call multiple times. No errors are returned.
func (s *Storage) Close(_ context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		s.db.Close()
	}
}
