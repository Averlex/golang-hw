package storage

import "errors"

var (
	// ErrNoData is returned when no data is passed to any of the CRUD methods.
	ErrNoData = errors.New("no data passed")
	// ErrTimeoutExceeded is returned when the operation execution times out.
	ErrTimeoutExceeded = errors.New("timeout exceeded")
	// ErrQeuryError is returned when the query execution fails.
	ErrQeuryError = errors.New("query execution")
	// ErrDataExists is returned on event ID collision on DB insertion.
	ErrDataExists = errors.New("event data already exists")
	// ErrNotExists is returned when the event with requested ID does not exist in the DB.
	ErrNotExists = errors.New("event does not exist")
)
