// Package storagetypes contains errors used in the storage package
// as well as Event type and its constructor and helper functions.
package storagetypes

import "errors"

// Storage operational errors.
var (
	// ErrNoData is returned when no data is passed to any of the CRUD methods.
	ErrNoData = errors.New("no data passed")
	// ErrTimeoutExceeded is returned when the operation execution times out.
	ErrTimeoutExceeded = errors.New("timeout exceeded")
	// ErrQeuryError is returned when the query execution fails.
	ErrQeuryError = errors.New("query execution")
	// ErrDataExists is returned on event ID collision on the storage insertion.
	ErrDataExists = errors.New("event data already exists")
	// ErrUnsupportedDriver is returned when the DB driver is not supported.
	ErrUnsupportedDriver = errors.New("unsupported driver, only 'postgres' and 'postgresql' are supported")
)

// Business errors.
var (
	// ErrEventNotFound is returned when the event with requested ID does not exist in the storage.
	ErrEventNotFound = errors.New("requested event was not found")
	// ErrDateBusy is returned when the event date is already busy/overlaps with existing events in the storage.
	ErrDateBusy = errors.New("requested event date is already busy")
	// ErrPermissionDenied is returned when the user tries to modify another user's event.
	ErrPermissionDenied = errors.New("cannot modify another user's event")
)

// Data validation errors.
var (
	// ErrEmptyField is returned when no data is passed to any of the necessary fields.
	ErrEmptyField = errors.New("empty event field values received")
	// ErrNegativeDuration is returned when the event duration is negative.
	ErrNegativeDuration = errors.New("event duration is negative")
	// ErrNegativeRemind is returned when the event remind duration is negative.
	ErrNegativeRemind = errors.New("event remind duration is negative")
	// ErrGenerateID is returned when the event ID generation fails.
	ErrGenerateID = errors.New("failed to generate new event id")
)
