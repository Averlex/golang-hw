package storage

import (
	"fmt"
	"reflect"
	"time"

	memorystorage "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory" //nolint:depguard,nolintlint
	sqlstorage "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"       //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors"                            //nolint:depguard,nolintlint
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

// newMemoryStorage creates a new memory storage instance.
// args is a map[string]any containing the configuration for the storage. All args are parsed and validated.
// Returns (nil, nil) or (*Storage, nil) if no errors occurred, (nil, error) otherwise.
func newMemoryStorage(args map[string]any) (Storage, error) {
	memArgs, ok := args["memory"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: no storage configuration received", errors.ErrCorruptedConfig)
	}

	missing, wrongType := validateMemoryConfig(memArgs)
	if len(missing) > 0 || len(wrongType) > 0 {
		return nil, fmt.Errorf("%w: missing=%v invalid_type=%v",
			errors.ErrCorruptedConfig, missing, wrongType)
	}

	size, _ := memArgs["size"].(int)
	size = max(0, size) // Ensure size is not negative.

	return memorystorage.NewStorage(size)
}

// newSQLStorage creates a new sql storage instance.
// args is a map[string]any containing the configuration for the storage. All args are parsed and validated.
// Returns (nil, nil) or (*Storage, nil) if no errors occurred, (nil, error) otherwise.
func newSQLStorage(args map[string]any) (Storage, error) {
	sqlArgs, ok := args["sql"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: no storage configuration received", errors.ErrCorruptedConfig)
	}

	missing, wrongType := validateSQLConfig(sqlArgs)
	if len(missing) > 0 || len(wrongType) > 0 {
		return nil, fmt.Errorf("%w: missing=%v invalid_type=%v",
			errors.ErrCorruptedConfig, missing, wrongType)
	}

	callArgs := map[string]string{
		"host":     sqlArgs["host"].(string),
		"port":     sqlArgs["port"].(string),
		"user":     sqlArgs["user"].(string),
		"password": sqlArgs["password"].(string),
		"dbname":   sqlArgs["dbname"].(string),
		"driver":   sqlArgs["driver"].(string),
	}

	timeout, _ := sqlArgs["timeout"].(time.Duration)
	timeout = max(0, timeout)

	return sqlstorage.NewStorage(
		timeout,
		callArgs["driver"],
		callArgs["host"],
		callArgs["port"],
		callArgs["user"],
		callArgs["password"],
		callArgs["dbname"],
	)
}
