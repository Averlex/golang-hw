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
//
// Currently supported drivers are "postgres" or "postgresql".
func NewStorage(timeout time.Duration, driver, host, port, user, password, dbname string) (*Storage, error) {
	if driver != "postgres" && driver != "postgresql" {
		return nil, sttypes.ErrUnsupportedDriver
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d",
		host, port, user, password, dbname, int(timeout.Seconds()),
	)

	return &Storage{
		driver:  driver,
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
// or pinging, it returns an error.
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
// If the connection is already closed or DB object is nil, it returns nil.
func (s *Storage) Close(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return nil
	}

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
