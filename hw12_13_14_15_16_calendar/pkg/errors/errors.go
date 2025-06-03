// Package errors contains errors used in the project.
package errors

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
	// ErrStorageUninitialized is returned when the database connection is not initialized.
	ErrStorageUninitialized = errors.New("storage is not initialized (initialize connection first?)")
	// ErrStorageFull is returned when the storage is full and cannot accept new events.
	ErrStorageFull = errors.New("storage is full")
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

// Setup errors.
var (
	// ErrCorruptedConfig is returned when config data is invalid or missing.
	ErrCorruptedConfig = errors.New("config data is invalid")
	// ErrStorageInitFailed is returned when the storage initialization fails.
	ErrStorageInitFailed = errors.New("storage initialization failed")
)

// Logger errors.
var (
	// ErrInvalidLogType  is an error that is returned when the log type is invalid.
	ErrInvalidLogType = errors.New("invalid log type")
	// ErrInvalidLogLevel is an error that is returned when the log level is invalid.
	ErrInvalidLogLevel = errors.New("invalid log level")
	// ErrInvalidWriter is an error that is returned when the writer is not set.
	ErrInvalidWriter = errors.New("invalid writer set")
	// ErrInvalidTimeTemplate is an error that is returned when the time template cannot be parsed by time package.
	ErrInvalidTimeTemplate = errors.New("invalid time template")
)
