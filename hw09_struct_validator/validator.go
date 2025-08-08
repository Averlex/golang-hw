// Package hw09structvalidator implements a struct validator and its error types.
package hw09structvalidator

import (
	"fmt"
	"reflect"
)

// Validate validates the given value.
func Validate(data any) error {
	if data == nil {
		return nil
	}

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", val.Kind().String())
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
			errs = errs.Add(field.Name, fmt.Errorf("empty tag"))
			continue
		}

		kind := field.Type.Kind()

		// Extracting concrete value and type if the field is an interface.
		if kind == reflect.Interface {
			kind = fieldVal.Type().Kind()
			fieldVal = reflect.ValueOf(fieldVal.Interface())
		}

		// If the field is a slice, check if it is a slice full of strings or integers.
		intVals := make([]int, 0)
		stringVals := make([]string, 0)
		isValidSlice := true
		switch kind {
		case reflect.Slice:
			count := 0
			for item := range fieldVal.Seq() {
				count++
				switch item.Type().Kind() {
				case reflect.Int:
					intVals = append(intVals, int(item.Int()))
				case reflect.String:
					stringVals = append(stringVals, item.String())
				}
			}
			if count != len(intVals) || count != len(stringVals) {
				isValidSlice = false
			}
		// For string and int types: perform the conversion T -> []T for helper compatability.
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
				errs = errs.Add(field.Name, fmt.Errorf("expected %q tag, got %q", nestedTag, tag))
				continue
			}
			nestedErrs = Validate(fieldVal.Interface())
			if len(nestedErrs) > 0 {
				errs = append(errs, nestedErrs...)
			}
		case isValidSlice && len(stringVals) > 0:
			stringErrs := validateStrings(stringVals, tag)
			if len(stringErrs) > 0 {
				errs = append(errs, stringErrs...)
			}
		case isValidSlice && len(intVals) > 0:
			intErrs := validateInts(intVals, tag)
			if len(intErrs) > 0 {
				errs = append(errs, intErrs...)
			}
		default:
			errs = errs.Add(field.Name, fmt.Errorf("unsupported type %s", kind.String()))
		}
	}

	return nil
}
