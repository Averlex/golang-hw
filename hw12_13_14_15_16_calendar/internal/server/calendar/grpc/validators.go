package grpc

import (
	"fmt"
	"reflect"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                          //nolint:depguard,nolintlint
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

// parseUUID parses a string into a UUID or returns an error if invalid.
func parseUUID(id string) (uuid.UUID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid id format in request data: %w", projectErrors.ErrInvalidFieldData, err)
	}
	return parsedID, nil
}
