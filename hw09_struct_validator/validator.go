// Package hw09structvalidator implements a struct validator and its error types.
package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// Validate validates the given value.
//
// Returns nil if the data is valid, or an error if it is not.
// The returned error type is always a ValidationErrors.
func Validate(data any) error {
	if data == nil {
		return nil
	}

	val := reflect.ValueOf(data)
	// Extract concrete value if data is an interface.
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("%w: expected struct, got %s", ErrInvalidData, val.Kind().String())
	}

	var errs ValidationErrors
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		if err := validateField(typ.Field(i), val.Field(i)); err != nil {
			if fieldErrs, critical, err := isCriticalError(err); critical {
				return fmt.Errorf("%w: field=%q", err, typ.Field(i).Name)
			} else if len(fieldErrs) > 0 {
				errs = append(errs, fieldErrs...)
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// validateField validates a single struct field based on its tag and type.
func validateField(field reflect.StructField, fieldVal reflect.Value) error {
	tag, ok := field.Tag.Lookup(lookupTag)
	// Skipping the field if it has no tag.
	if !ok {
		return nil
	}
	// Empty tag is not valid.
	if tag == "" {
		return fmt.Errorf("%w: received an empty tag: field=%q", ErrInvalidData, field.Name)
	}

	kind := field.Type.Kind()
	// Extracting concrete value and type if the field is an interface.
	if kind == reflect.Interface {
		fieldVal = fieldVal.Elem()
		kind = fieldVal.Type().Kind()
	}

	//nolint:exhaustive
	switch kind {
	case reflect.Struct:
		return validateNestedStruct(fieldVal, field.Name, tag)
	case reflect.Slice:
		return validateSlice(fieldVal, field.Name, tag)
	case reflect.String:
		return validateStrings([]string{fieldVal.String()}, field.Name, tag)
	case reflect.Int:
		return validateInts([]int{int(fieldVal.Int())}, field.Name, tag)
	default:
		return fmt.Errorf("%w: unsupported type %s: field=%q", ErrInvalidData, kind.String(), field.Name)
	}
}

// validateNestedStruct validates a nested struct field.
func validateNestedStruct(fieldVal reflect.Value, fieldName, tag string) error {
	if tag != nestedTag {
		return fmt.Errorf("%w: expected %q tag, got %q: field=%q", ErrInvalidData, nestedTag, tag, fieldName)
	}
	return Validate(fieldVal.Interface())
}

// validateSlice validates a slice field of strings or integers.
func validateSlice(fieldVal reflect.Value, fieldName, tag string) error {
	intVals, stringVals, err := extractSliceValues(fieldVal)
	if err != nil {
		return fmt.Errorf("%w: field=%q", err, fieldName)
	}

	if len(intVals) > 0 {
		return validateInts(intVals, fieldName, tag)
	}
	if len(stringVals) > 0 {
		return validateStrings(stringVals, fieldName, tag)
	}
	// Empty slice validation: try both int and string validations.
	intErr := validateInts(intVals, fieldName, tag)
	stringErr := validateStrings(stringVals, fieldName, tag)
	if intErr != nil && stringErr != nil {
		return fmt.Errorf(
			"%w: unable to detect type %s: field=%q",
			ErrInvalidData, fieldVal.Type().Kind().String(), fieldName,
		)
	}
	return nil
}

// extractSliceValues extracts integer or string values from a slice.
func extractSliceValues(fieldVal reflect.Value) ([]int, []string, error) {
	intVals := make([]int, 0)
	stringVals := make([]string, 0)
	for i := 0; i < fieldVal.Len(); i++ {
		elem := fieldVal.Index(i)
		elemKind := elem.Type().Kind()
		if elemKind == reflect.Interface {
			elem = elem.Elem()
			elemKind = elem.Type().Kind()
		}
		//nolint:exhaustive
		switch elemKind {
		case reflect.Int:
			intVals = append(intVals, int(elem.Int()))
		case reflect.String:
			stringVals = append(stringVals, elem.String())
		default:
			return nil, nil, fmt.Errorf("%w: unsupported slice element type %s", ErrInvalidData, elemKind.String())
		}
	}
	if len(intVals) > 0 && len(stringVals) > 0 {
		return nil, nil, fmt.Errorf("%w: mixed types in slice", ErrInvalidData)
	}
	return intVals, stringVals, nil
}

// isCriticalError checks if the error is critical and returns (err, nil, true) if so.
//
// If the error is not critical, returns (nil, errs, false).
func isCriticalError(err error) (ValidationErrors, bool, error) {
	if err == nil {
		return nil, false, nil
	}

	// Not recoverable program error occurred somewhere during validation.
	if errors.Is(err, ErrInvalidData) {
		return nil, true, err
	}
	// Expecting ValidationErrors otherwise.
	var errs ValidationErrors
	ok := errors.As(err, &errs)
	if !ok {
		return nil, true, fmt.Errorf("%w: expected <ValidationErrors> type as an error, got %T", ErrInvalidData, err)
	}
	return errs, false, nil
}

// validateCommands checks if the commands are valid for a given field.
// This helper does not checks if the instructions are valid.
func validateCommands(tag string, expectedCommands []string) (map[string]string, error) {
	if tag == "" {
		return nil, fmt.Errorf("%w: received an empty tag", ErrInvalidData)
	}

	// Verifying the number of commands for a single field.
	commands := strings.Split(tag, "|")
	if len(commands) > tagLimit {
		return nil, fmt.Errorf("%w: received %d commands, expected [1, %d]", ErrInvalidData, len(commands), tagLimit)
	}

	cmd := make(map[string]string)
	for _, t := range commands {
		// Splitting the tag into a command and its instruction.
		cmdData := strings.Split(t, ":")
		if len(cmdData) != cmdPartsNumber {
			return nil, fmt.Errorf("%w: incorrect command format, expected <command:instruction>, got <%s>", ErrInvalidData, t)
		}
		if !slices.Contains(expectedCommands, cmdData[0]) {
			return nil, fmt.Errorf("%w: unknown command %q", ErrInvalidData, cmdData[0])
		}
		if _, ok := cmd[cmdData[0]]; ok {
			return nil, fmt.Errorf("%w: duplicate command %q", ErrInvalidData, cmdData[0])
		}

		cmd[cmdData[0]] = cmdData[1]
	}

	return cmd, nil
}
