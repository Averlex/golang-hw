package app

import (
	"errors"
	"reflect"

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

// safeDereference returns zero value if ptr is nil.
func safeDereference[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}
