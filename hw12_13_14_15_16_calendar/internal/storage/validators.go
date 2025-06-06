package storage

import (
	"reflect"
	"time"
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

// validateSQLConfig returns missing and wrong type fields of sql config found in args.
func validateSQLConfig(args map[string]any) ([]string, []string) {
	required := map[string]any{
		"driver":   "",
		"host":     "",
		"port":     "",
		"user":     "",
		"password": "",
		"dbname":   "",
		"timeout":  time.Duration(0),
	}

	return validateFields(args, required)
}

// validateMemoryConfig returns missing and wrong type fields memory config found in args.
func validateMemoryConfig(args map[string]any) ([]string, []string) {
	required := map[string]any{
		"size": int(0),
	}

	return validateFields(args, required)
}
