// Package sql provides a SQL database storage implementation.
package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
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

// withTimeout wraps the given function in a context.WithTimeout call.
func (s *Storage) withTimeout(ctx context.Context, fn func(context.Context) error) error {
	s.mu.RLock()
	if s.db == nil {
		return projectErrors.ErrStorageUninitialized
	}

	timeout := s.timeout
	s.mu.RUnlock()

	if timeout == 0 {
		return fn(ctx)
	}

	localCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := fn(localCtx)
	if err != nil {
		if errors.Is(localCtx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("%w: %w", projectErrors.ErrTimeoutExceeded, err)
		}
		return err
	}
	return nil
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
			return fmt.Errorf("database connection: %w", err)
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

// execInTransaction executes the given function in a transaction.
//
// If the function returns an error, the transaction is rolled back and the error
// is returned. If the function succeeds, the transaction is committed and
// any error that occurs during the commit is returned after the rollback.
func (s *Storage) execInTransaction(ctx context.Context, fn func(context.Context, Tx) error) error {
	if s.db == nil {
		return projectErrors.ErrStorageUninitialized
	}

	return s.withTimeout(ctx, func(localCtx context.Context) error {
		tx, err := s.db.BeginTxx(localCtx, nil)
		if err != nil {
			return fmt.Errorf("transaction begin: %w", err)
		}

		// This code wraps the original error if it fails to rollback the transaction.
		defer func() {
			// Rollback the transaction in any case except the one without any errors.
			if err != nil {
				rErr := tx.Rollback()
				if rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
					rErr = fmt.Errorf("transaction rollback: %w", rErr)
					err = fmt.Errorf("%w: %w", rErr, err)
				}
			}
		}()

		err = fn(localCtx, tx)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("transaction commit: %w", err)
		}

		return nil
	})
}
