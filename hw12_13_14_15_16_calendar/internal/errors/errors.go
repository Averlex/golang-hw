// Package errors contains errors used in the project.
package errors

import "errors"

// Setup errors.
// CMD level.
// Fatal.
var (
	// ErrCorruptedConfig is returned when config data is invalid or missing.
	ErrCorruptedConfig = errors.New("config data is invalid")
	// ErrStorageInitFailed is returned when the storage initialization fails.
	ErrStorageInitFailed = errors.New("storage initialization failed")
	// ErrLoggerInitFailed is returned when the logger initialization fails.
	ErrLoggerInitFailed = errors.New("logger initialization failed")
	// ErrAppInitFailed is returned when the app initialization fails.
	ErrAppInitFailed = errors.New("app initialization failed")
	// ErrServerInitFailed is returned when the server initialization fails.
	ErrServerInitFailed = errors.New("server initialization failed")
	// ErrUnsupportedDriver is returned when the DB driver is not supported.
	ErrUnsupportedDriver = errors.New("unsupported driver, only 'postgres' and 'postgresql' are supported")
)

// Storage operational errors - critical.
// Server level.
// ERROR.
var (
	// ErrStorageFull is returned when the storage is full and cannot accept new events.
	ErrStorageFull = errors.New("storage is full")
)

// Business errors.
// Server level.
// INFO - potential 40x codes.
var (
	// ErrEventNotFound is returned when the event with requested ID does not exist in the storage.
	ErrEventNotFound = errors.New("requested event was not found")
	// ErrDateBusy is returned when the event date is already busy/overlaps with existing events in the storage.
	ErrDateBusy = errors.New("requested event date is already busy")
	// ErrPermissionDenied is returned when the user tries to modify another user's event.
	ErrPermissionDenied = errors.New("cannot modify another user's event")
	// ErrNoData is returned when no data is passed to any of the CRUD methods.
	ErrNoData = errors.New("no data passed")
)

// Data validation errors.
// Server level.
// INFO - potential 40x codes.
var (
	// ErrEmptyField is returned when no data is passed to any of the necessary fields.
	ErrEmptyField = errors.New("empty event field values received")
	// ErrIvalidFieldData is returned when invalid data is passed to any of the fields.
	ErrInvalidFieldData = errors.New("invalid event field values received")
)

// Unsuccessful result of storage operations.
// App and Server level.
// ERROR on App, WARN on Server.
var (
	// ErrRetriesExceeded is returned when the retry count is exceeded.
	ErrRetriesExceeded = errors.New("maximum retries exceeded")
)

// Errors, which breaks the normal execution flow.
// App level. Not retryable.
// ERROR.
var (
	// ErrInconsistentState is returned when an unexpected internal error occurs.
	ErrInconsistentState = errors.New("unexpected internal error")
)

// Storage errors.
// App level, retryable.
// ERROR.
var (
	// ErrStorageUninitialized is returned when the database connection is not initialized.
	ErrStorageUninitialized = errors.New("storage is not initialized (initialize connection first?)")
)

// Storage operational errors - retryable.
// App level until the retry count is exceeded.
// DEBUG.
var (
	// ErrTimeoutExceeded is returned when the operation execution times out.
	ErrTimeoutExceeded = errors.New("timeout exceeded")
	// ErrQeuryError is returned when the query execution fails.
	ErrQeuryError = errors.New("query execution")
	// ErrDataExists is returned on event ID collision on the storage insertion.
	ErrDataExists = errors.New("event data already exists")
	// ErrGenerateID is returned when the event ID generation fails.
	ErrGenerateID = errors.New("failed to generate new event id")
)
