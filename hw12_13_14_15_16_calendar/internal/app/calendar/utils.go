package app

import (
	"errors"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
)

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
