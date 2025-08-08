// Package hw09structvalidator implements a struct validator and its error types.
package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
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
	typ := val.Type()

	errs := make(ValidationErrors, 0)
	for i := range typ.NumField() {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tag, ok := field.Tag.Lookup(lookupTag)
		// Skipping the field if it has no tag.
		if !ok {
			continue
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

		// If the field is a slice, check if it is a slice full of strings or integers.
		intVals := make([]int, 0)
		stringVals := make([]string, 0)
		isValidSlice := true
		//nolint:exhaustive
		switch kind {
		case reflect.Slice:
			count := 0
			for i := range fieldVal.Len() {
				count++
				elem := fieldVal.Index(i)
				elemKind := elem.Type().Kind()
				if elemKind == reflect.Interface {
					elem = elem.Elem()
					elemKind = elem.Type().Kind()
				}
				switch elemKind {
				case reflect.Int:
					intVals = append(intVals, int(elem.Int()))
				case reflect.String:
					stringVals = append(stringVals, elem.String())
				}
			}
			if count != len(intVals) && count != len(stringVals) {
				isValidSlice = false
			}
		// For string and int types: perform the conversion T -> []T for helper compatibility.
		case reflect.String:
			stringVals = append(stringVals, fieldVal.String())
		case reflect.Int:
			intVals = append(intVals, int(fieldVal.Int()))
		}

		// Perform the final kind check.
		switch {
		case kind == reflect.Struct:
			// We have a nested struct with 'validate' tag, but it's value is not 'nested'.
			if tag != nestedTag {
				return fmt.Errorf("%w: expected %q tag, got %q: field=%q", ErrInvalidData, nestedTag, tag, field.Name)
			}
			err := Validate(fieldVal.Interface())
			nestedErrs, isCritical, err := isCriticalError(err)
			if isCritical {
				return fmt.Errorf("%w: field=%q", err, field.Name)
			}
			if len(nestedErrs) > 0 {
				errs = append(errs, nestedErrs...)
			}
		case isValidSlice && len(stringVals) > 0:
			err := validateStrings(stringVals, field.Name, tag)
			stringErrs, isCritical, err := isCriticalError(err)
			if isCritical {
				return err
			}
			if len(stringErrs) > 0 {
				errs = append(errs, stringErrs...)
			}
		case isValidSlice && len(intVals) > 0:
			err := validateInts(intVals, field.Name, tag)
			intErrs, isCritical, err := isCriticalError(err)
			if isCritical {
				return fmt.Errorf("%w: field=%q", err, field.Name)
			}
			if len(intErrs) > 0 {
				errs = append(errs, intErrs...)
			}
		// The field is an empty slice, so it should be valid either as []int or []string.
		// If it is not, then it is an error.
		case isValidSlice && len(stringVals) == 0 && len(intVals) == 0:
			intErr := validateInts(intVals, field.Name, tag)
			stringErr := validateStrings(stringVals, field.Name, tag)
			if intErr != nil && stringErr != nil {
				return fmt.Errorf("%w: unable to detect type %s: field=%q", ErrInvalidData, kind.String(), field.Name)
			}
			return nil
		default:
			return fmt.Errorf("%w: unsupported type %s: field=%q", ErrInvalidData, kind.String(), field.Name)
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
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

// validateStrings checks if the commands and instructions for the tag of the field.
// If any of the commands or instructions is invalid, returns a wrapped ErrInvalidData error.
//
// Returns nil if the data is valid or ValidationErrors if not.
//
// Empty slice is considered a valid one.
func validateStrings(vals []string, fieldName, tag string) error {
	errs := make(ValidationErrors, 0)

	cmd, err := validateCommands(tag, stringCommands)
	if err != nil {
		return fmt.Errorf("%w: field=%q", err, fieldName)
	}

	for k, v := range cmd {
		switch k {
		case lenCmd:
			instructions := strings.Split(v, ",")
			if len(instructions) != 1 {
				return fmt.Errorf("%w: expected 1 instruction for tag %q, got %d: field=%q",
					ErrInvalidData, k, len(instructions), fieldName,
				)
			}
			length, err := strconv.Atoi(instructions[0])
			if err != nil {
				return fmt.Errorf("%w: expected integer value for tag %q, got %s: field=%q",
					ErrInvalidData, k, instructions[0], fieldName,
				)
			}
			for i, s := range vals {
				if len(s) != length {
					errs = errs.Add(
						fieldName,
						fmt.Errorf("string #%d length doesn't meet the length requirement: expected %d, got %d", i, length, len(s)))
				}
			}
		case regexpCmd:
			expression, err := regexp.Compile(v)
			if err != nil {
				return fmt.Errorf("%w: expected valid regular expression for tag %q, got %s: field=%q",
					ErrInvalidData, k, v, fieldName,
				)
			}
			for i, s := range vals {
				if !expression.MatchString(s) {
					errs = errs.Add(
						fieldName,
						fmt.Errorf("string #%d doesn't meet the regexp requirement: expected %s, got %s", i, v, s))
				}
			}
		case inCmd:
			instructions := strings.Split(v, ",")
			if len(instructions) < 1 {
				return fmt.Errorf("%w: expected >=1 instructions for tag %q, got %d: field=%q",
					ErrInvalidData, k, len(instructions), fieldName,
				)
			}
			for i, s := range vals {
				if !slices.Contains(instructions, s) {
					errs = errs.Add(
						fieldName,
						fmt.Errorf("string #%d doesn't meet the slice requirement: expected %s, got %s", i, v, s))
				}
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func validateInts(vals []int, fieldName, tag string) error {
	errs := make(ValidationErrors, 0)

	cmd, err := validateCommands(tag, intCommands)
	if err != nil {
		return fmt.Errorf("%w: field=%q", err, fieldName)
	}

	for k, v := range cmd {
		switch k {
		case minCmd, maxCmd:
			instructions := strings.Split(v, ",")
			if len(instructions) != 1 {
				return fmt.Errorf("%w: expected 1 instruction for tag %q, got %d: field=%q",
					ErrInvalidData, k, len(instructions), fieldName,
				)
			}
			val, err := strconv.Atoi(instructions[0])
			if err != nil {
				return fmt.Errorf("%w: expected integer value for tag %q, got %s: field=%q",
					ErrInvalidData, k, instructions[0], fieldName,
				)
			}
			for i, d := range vals {
				var expectedSign string
				if k == minCmd && d < val {
					expectedSign = ">="
				} else if k == maxCmd && d > val {
					expectedSign = "<="
				}
				if expectedSign != "" {
					errs = errs.Add(
						fieldName,
						fmt.Errorf("int #%d value violates the %s requirement: expected %s%d, got %d", i, k, expectedSign, val, d))
				}
			}
		case inCmd:
			instructions := strings.Split(v, ",")
			if len(instructions) < 1 {
				return fmt.Errorf("%w: expected >=1 instructions for tag %q, got %d: field=%q",
					ErrInvalidData, k, len(instructions), fieldName,
				)
			}
			intInstuctions := make([]int, len(instructions))
			for i, s := range instructions {
				intInstuctions[i], err = strconv.Atoi(s)
				if err != nil {
					return fmt.Errorf("%w: expected integer value for #%d tag %q, got %s: field=%q",
						ErrInvalidData, i, k, s, fieldName,
					)
				}
			}
			for i, d := range vals {
				if !slices.Contains(intInstuctions, d) {
					errs = errs.Add(
						fieldName,
						fmt.Errorf("int #%d doesn't meet the slice requirement: expected %s, got %d", i, v, d))
				}
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

// validateCommands checks if the commands are valid for a given field.
// This helper does not checks if the instructions are valid.
func validateCommands(tag string, expectedCommands []string) (map[string]string, error) {
	if tag == "" {
		return nil, fmt.Errorf("%w: received an empty tag", ErrInvalidData)
	}

	// Verifying the number of tags for a single field.
	tags := strings.Split(tag, "|")
	if len(tags) > tagLimit {
		return nil, fmt.Errorf("%w: received %d tags, expected [1, %d]", ErrInvalidData, len(tags), tagLimit)
	}

	cmd := make(map[string]string)
	for _, t := range tags {
		// Splitting the tag into a command and its instruction.
		cmdData := strings.Split(t, ":")
		if len(cmdData) != cmdPartsNumber {
			return nil, fmt.Errorf("%w: incorrect tag format, expected <command:instruction>, got <%s>", ErrInvalidData, t)
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
