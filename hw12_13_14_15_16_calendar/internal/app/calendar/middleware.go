package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
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
func (a *App) withRetries(ctx context.Context, method string, fn func() error) error {
	var err error
	a.mu.RLock()
	attempts := a.retries + 1 // To guarantee at least 1 execution.
	timeout := a.retryTimeout
	a.mu.RUnlock()

	msg := "operation failed"

	for i := range attempts {
		err = fn()
		if err == nil {
			return nil
		}

		if !a.isRetryable(err) {
			// The error here is not really expected, therefore no special processing implemented.
			if !a.isUninitialized(err) {
				a.l.Error(ctx,
					"unexpected error occurred on method call",
					slog.String("method", method),
					slog.Any("error", err),
				)
				return fmt.Errorf(msg+": %w", err)
			}
			// ErrStorageUninitialized is really retryable but with extra logic and another logging level.
			a.l.Error(ctx,
				"operation failed due to storage unavailability",
				slog.String("method", method),
				slog.Int("attempt", i+1),
				slog.Any("error", err),
			)
			a.l.Info(ctx, "attempting to connect to storage", slog.String("method", method), slog.Int("attempt", i+1))
			err = a.s.Connect(ctx)
			if err != nil {
				a.l.Error(ctx,
					"unable re-establish storage connection",
					slog.String("method", method),
					slog.Int("attempt", i+1),
					slog.Any("error", err),
				)
				return fmt.Errorf(msg+": %w", err)
			}
			a.l.Info(ctx, "storage connection re-established", slog.String("method", method), slog.Int("attempt", i+1))
		} else {
			a.l.Debug(ctx, "operation failed", slog.String("method", method), slog.Int("attempt", i+1), slog.Any("error", err))
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
		a.l.Error(ctx,
			"retry limit exceeded",
			slog.String("method", method),
			slog.Int("total_attempts", attempts),
			slog.Any("error", err),
		)
		return fmt.Errorf("%w: %w", projectErrors.ErrRetriesExceeded, err)
	}

	return nil
}
