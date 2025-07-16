package memory

import (
	"context"
	"fmt"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
)

// withLockAndChecks is a helper function that performs common checks and locking for storage operations.
// It checks if the storage is initialized, acquires a lock, executes the provided function.
//
// If any afterCtx function is provided, it will be called after context is checked (same behavior as tx.commit).
//
// If any rollback function is provided, it will be called in case of an error during the operation or due to timeout.
//
// Both rollback and afterCtx functions are optional and can be nil. They also should not return any errors and panic.
func (s *Storage) withLockAndChecks(ctx context.Context,
	beforeCtx func() error, afterCtx, rollback func(),
	muMode mutexMode,
) error {
	// Acquire lock.
	if muMode == writeLock {
		s.mu.Lock()
		defer s.mu.Unlock()
	} else {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}

	// Check storage init.
	if err := s.checkState(); err != nil {
		return err
	}

	if beforeCtx == nil {
		return fmt.Errorf("no action provided to execute")
	}

	// Execute prepared operation.
	err := beforeCtx()
	if err != nil {
		// Trying to rollback changes if rollback function is provided.
		if rollback != nil {
			rollback()
		}
		return err
	}

	// Check context before applying changes.
	if err := ctx.Err(); err != nil {
		// Trying to rollback changes if rollback function is provided.
		if rollback != nil {
			rollback()
		}
		return fmt.Errorf("%w: %w", projectErrors.ErrTimeoutExceeded, err)
	}

	// Executing final actions.
	if afterCtx != nil {
		afterCtx()
	}

	return nil
}
