package hw09structvalidator

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError is an error type for validation errors.
type ValidationError struct {
	Field string
	Err   error
}

// ValidationErrors implements MultiError pattern for ValidationError.
type ValidationErrors []ValidationError

// Unwrap provides errors.As() support for ValidationErrors.
func (v ValidationErrors) Unwrap() []error {
	var errs []error
	for _, e := range v {
		if e.Err != nil {
			errs = append(errs, e.Err)
		}
	}
	return errs
}

// Is implements errors.Is() support for ValidationErrors.
func (v ValidationErrors) Is(target error) bool {
	for _, e := range v {
		if errors.Is(e.Err, target) {
			return true
		}
	}
	return false
}

// Error implements error interface according to MultiError pattern.
// It returns the combined message of all errors in the slice.
func (v ValidationErrors) Error() string {
	messages := make([]string, 0, len(v))
	for _, err := range v {
		// Skip nil errors.
		if err.Err == nil {
			continue
		}
		messages = append(messages, fmt.Sprintf("field=\"%s\", error=\"%s\"", err.Field, err.Err.Error()))
	}
	if len(messages) == 0 {
		return ""
	}

	return strings.Join(messages, "; ")
}

// Add adds a new ValidationError to the slice. Returns a new slice instead of modifying the original.
func (v ValidationErrors) Add(field string, err error) ValidationErrors {
	if err != nil {
		v = append(v, ValidationError{Field: field, Err: err})
	}
	return v
}

// NewValidationError returns a slice with a single ValidationError.
func NewValidationError(field string, err error) ValidationErrors {
	return ValidationErrors{
		{Field: field, Err: err},
	}
}
