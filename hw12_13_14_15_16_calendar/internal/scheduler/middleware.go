package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
)

// withRetries is an app middleware that implemetes the retry logic.
// It is trying to execute the given function with given retries and timeout, executing it at least once.
//
// If an error occurs during the execution and it is retryable, it will be retried.
// Each attempt is logged with DEBUG level.
//
// If the the storage connection is unavailable for some reason, the method will try to reconnect.
// Each fact of unavailability is logged with ERROR level, each reconnection attempt - with INFO.
//
// Receiving any other error stops the execution and passes the error on the upper level.
//
// If the retry limit is exceeded, ErrRetriesExceeded is returned, wrapped over the last occurred error.
func (sch *Scheduler) withRetries(ctx context.Context, method string, fn func() error) error {
	var err error
	sch.mu.RLock()
	attempts := sch.retries + 1 // To guarantee at least 1 execution.
	timeout := sch.retryTimeout
	sch.mu.RUnlock()

	msg := "operation failed"

	for i := range attempts {
		err = fn()
		if err == nil {
			return nil
		}

		if !sch.isRetryable(err) {
			// The error here is not really expected, therefore no special processing implemented.
			if !sch.isUninitialized(err) && !sch.isBusiness(err) {
				// Logging only really unexpected errors.
				sch.l.Error(ctx,
					"unexpected error occurred on method call",
					slog.String("method", method),
					slog.Any("error", err),
				)
			}
			// Both business and unexpected errors are not retryable.
			if !sch.isUninitialized(err) {
				return fmt.Errorf(msg+": %w", err)
			}
			// ErrStorageUninitialized is really retryable but with extra logic and another logging level.
			sch.l.Error(ctx,
				"operation failed due to storage unavailability",
				slog.String("method", method),
				slog.Int("attempt", i+1),
				slog.Any("error", err),
			)
			sch.l.Info(ctx, "attempting to connect to storage", slog.String("method", method), slog.Int("attempt", i+1))
			err = sch.s.Connect(ctx)
			if err != nil {
				sch.l.Error(ctx,
					"unable re-establish storage connection",
					slog.String("method", method),
					slog.Int("attempt", i+1),
					slog.Any("error", err),
				)
				return fmt.Errorf(msg+": %w", err)
			}
			sch.l.Info(ctx, "storage connection re-established", slog.String("method", method), slog.Int("attempt", i+1))
		} else {
			sch.l.Debug(ctx, "operation failed", slog.String("method", method), slog.Int("attempt", i+1), slog.Any("error", err))
		}

		// No need in waiting on the last attempt.
		if i < attempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(timeout):
				continue
			}
		}
	}

	if err != nil {
		sch.l.Error(ctx,
			"retry limit exceeded",
			slog.String("method", method),
			slog.Int("total_attempts", attempts),
			slog.Any("error", err),
		)
		return fmt.Errorf("%w: %w", projectErrors.ErrRetriesExceeded, err)
	}

	return nil
}
