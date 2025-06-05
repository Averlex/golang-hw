package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
)

// validateFields returns missing and wrong type fields found in args.
// requiredFields is a map of field names with their expected types.
func validateFields(args map[string]any, requiredFields map[string]any) ([]string, []string) {
	var missing []string
	var wrongType []string

	for field, expectedVal := range requiredFields {
		val, exists := args[field]
		if !exists {
			missing = append(missing, field)
			continue
		}

		expectedReflect := reflect.TypeOf(expectedVal)
		valueReflect := reflect.TypeOf(val)

		// Default type switch will end up with false positive results.
		// E.g., 123.(string) -> ok.
		if expectedReflect != valueReflect {
			wrongType = append(wrongType, field)
		}
	}

	return missing, wrongType
}

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
func (a *App) withRetries(ctx context.Context, fn func() error) error {
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
			if !a.isUninitialized(err) {
				return fmt.Errorf(msg+": %w", err)
			}
			// ErrStorageUninitialized is really retryable but with extra logic and another logging level.
			a.l.Error(ctx, msg, slog.Int("attempt", i+1), slog.Any("error", err))
			a.l.Info(ctx, "attempting to connect to storage", slog.Int("attempt", i+1))
			err = a.s.Connect(ctx)
			if err != nil {
				return fmt.Errorf(msg+": %w", err)
			}
		} else {
			a.l.Debug(ctx, msg, slog.Int("attempt", i+1), slog.Any("error", err))
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
		return fmt.Errorf("%w: %w", projectErrors.ErrRetriesExceeded, err)
	}

	return nil
}

func (a *App) isRetryable(err error) bool {
	return errors.Is(err, projectErrors.ErrTimeoutExceeded) ||
		errors.Is(err, projectErrors.ErrQeuryError) ||
		errors.Is(err, projectErrors.ErrDataExists) ||
		errors.Is(err, projectErrors.ErrGenerateID)
}

func (a *App) isUninitialized(err error) bool {
	return errors.Is(err, projectErrors.ErrStorageUninitialized)
}

// safeDereference returns zero value if ptr is nil.
func safeDereference[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}
