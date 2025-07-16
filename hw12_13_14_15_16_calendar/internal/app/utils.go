package app

import (
	"errors"
	"fmt"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                               //nolint:depguard,nolintlint
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

func (a *App) isBusiness(err error) bool {
	return errors.Is(err, projectErrors.ErrDateBusy) ||
		errors.Is(err, projectErrors.ErrPermissionDenied) ||
		errors.Is(err, projectErrors.ErrEventNotFound) ||
		errors.Is(err, projectErrors.ErrNoData)
}

// safeDereference returns zero value if ptr is nil.
func safeDereference[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

func idFromString(id string) (*uuid.UUID, error) {
	res, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w: event id: %w", projectErrors.ErrInvalidFieldData, err)
	}
	return &res, nil
}
