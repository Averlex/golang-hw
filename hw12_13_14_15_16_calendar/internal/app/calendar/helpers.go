package app

import (
	"context"
	"errors"
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

func (a *App) withRetries(ctx context.Context, fn func() error) error {
	var err error
	a.mu.RLock()
	retries := a.retries
	timeout := a.retryTimeout
	a.mu.RUnlock()

	// To guarantee at least 1 execution.
	retries++

	var i int
	for i = range retries {
		err = fn()
		if err == nil {
			return nil
		}

		if !a.isRetryable(err) {
			if !a.isUninitialized(err) {
				return err
			}
		}
		// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		// ВОТ ТУТ ВСЁ ОСТАНОВИЛОСЬЬЬЬЬЬЬЬЬЬЬЬЬЬЬЬЬ!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

		a.l.Debug(ctx, "operation failed", slog.Int("attempt", i+1), slog.Any("error", err))

		// No need in waiting on the last attempt.
		if i < retries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(timeout):
				continue
			}
		}
	}
	return err
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
