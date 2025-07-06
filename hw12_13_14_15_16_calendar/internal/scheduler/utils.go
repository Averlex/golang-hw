package scheduler

import (
	"errors"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
)

func (sch *Scheduler) isRetryable(err error) bool {
	return errors.Is(err, projectErrors.ErrTimeoutExceeded) ||
		errors.Is(err, projectErrors.ErrQeuryError)
}

func (sch *Scheduler) isUninitialized(err error) bool {
	return errors.Is(err, projectErrors.ErrStorageUninitialized)
}

func (sch *Scheduler) isBusiness(err error) bool {
	return errors.Is(err, projectErrors.ErrEventNotFound) ||
		errors.Is(err, projectErrors.ErrNoData)
}
