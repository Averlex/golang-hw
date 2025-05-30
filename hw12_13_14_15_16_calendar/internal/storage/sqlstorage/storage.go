// Package sqlstorage provides a SQL database storage implementation.
package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	sttypes "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/storagetypes" //nolint:depguard,nolintlint
	//nolint:depguard,nolintlint
	"github.com/jmoiron/sqlx" //nolint:depguard,nolintlint
)

const defaultDriver = "postgres"

// Storage represents a SQL database storage.
type Storage struct {
	mu      sync.RWMutex
	driver  string
	db      *sqlx.DB
	dsn     string
	timeout time.Duration
}

// NewStorage creates a new Storage instance based on the given args.
//
// If the arguments are empty, it returns an error.
//
// The function constructs a DSN based on the given arguments and
// the default driver. No connection is established upon the call.
func NewStorage(timeout time.Duration, host, port, user, password, dbname string) (*Storage, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d",
		host, port, user, password, dbname, int(timeout.Seconds()),
	)

	return &Storage{
		driver:  defaultDriver,
		dsn:     dsn,
		timeout: timeout,
	}, nil
}

// withTimeout wraps the given function in a context.WithTimeout call.
func (s *Storage) withTimeout(ctx context.Context, fn func(context.Context) error) error {
	s.mu.RLock()
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
			return fmt.Errorf("%w: %w", sttypes.ErrTimeoutExceeded, err)
		}
		return err
	}
	return nil
}

// Connect connects to the database.
//
// If the connection is successful, it pings the database
// to check if the connection is alive. If any error occurs during the connection
// or pinging, it returns an error with ErrDBConnection wrapped around the
// original error.
func (s *Storage) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.withTimeout(ctx, func(localCtx context.Context) error {
		var err error
		s.db, err = sqlx.ConnectContext(localCtx, s.driver, s.dsn)
		if err != nil {
			return fmt.Errorf("database connection: %w", err)
		}
		return nil
	})
}

// Close closes the connection to the database.
//
// It pings the database to check if the connection is alive before closing it.
// If any error occurs during the connection or pinging, it returns an error with
// ErrDBConnection wrapped around the original error.
func (s *Storage) Close(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.Close()
	return nil
}

// execInTransaction executes the given function in a transaction.
//
// If the function returns an error, the transaction is rolled back and the error
// is returned. If the function succeeds, the transaction is committed and
// any error that occurs during the commit is returned after the rollback.
func (s *Storage) execInTransaction(ctx context.Context, fn func(context.Context, *sqlx.Tx) error) error {
	return s.withTimeout(ctx, func(localCtx context.Context) error {
		tx, err := s.db.BeginTxx(localCtx, nil)
		if err != nil {
			return fmt.Errorf("transaction begin: %w", err)
		}

		// To ensure we don't mute the possible error.
		var rollbackErr error
		defer func() {
			if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
				rollbackErr = fmt.Errorf("transaction rollback: %w", err)
			}
		}()

		if err := fn(localCtx, tx); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("transaction commit: %w", err)
		}
		return rollbackErr
	})
}
