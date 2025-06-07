package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
)

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

// execInTransaction executes the given function in a transaction.
//
// If the function returns an error, the transaction is rolled back and the error
// is returned. If the function succeeds, the transaction is committed and
// any error that occurs during the commit is returned after the rollback.
func (s *Storage) execInTransaction(ctx context.Context, fn func(context.Context, Tx) error) error {
	s.mu.RLock()
	if s.db == nil {
		return projectErrors.ErrStorageUninitialized
	}
	s.mu.RUnlock()

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
