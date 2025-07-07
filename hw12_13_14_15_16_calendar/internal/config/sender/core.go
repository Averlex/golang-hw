package sender

import (
	"fmt"
	"reflect"
	"time"
)

// GetSubConfig returns a nested section of the configuration as a map[string]any.
// The section is identified by the given key, which matches either the field name or its mapstructure tag.
// If the key does not correspond to a struct field, an error is returned.
func (c *Config) GetSubConfig(key string) (map[string]any, error) {
	val := reflect.ValueOf(c).Elem()
	typ := val.Type()

	for i := range typ.NumField() {
		field := typ.Field(i)
		tag := field.Tag.Get("mapstructure")

		// Do not compare with the field name itself, as it is CamelCased by default.
		if tag == key {
			fieldVal := val.Field(i)

			if fieldVal.Kind() != reflect.Struct {
				return nil, fmt.Errorf("subsection %q is not a subconfig", key)
			}

			return structToMap(fieldVal), nil
		}
	}

	return nil, fmt.Errorf("subsection %q not found", key)
}

// structToMap recursively converts a struct value into a map[string]any.
// It supports nested structs and handles time.Duration fields by converting them to their string representation.
// All other fields are added as-is.
func structToMap(v reflect.Value) map[string]any {
	res := make(map[string]any)

	typ := v.Type()
	for i := range typ.NumField() {
		field := typ.Field(i)
		tag := field.Tag.Get("mapstructure")

		name := tag
		if name == "" {
			name = field.Name
		}

		value := v.Field(i)

		//nolint:exhaustive
		switch value.Kind() {
		// Expect time.Duration, string or struct fields.
		case reflect.Struct:
			if field.Type == reflect.TypeOf(time.Duration(0)) {
				res[name] = value.Interface()
			} else {
				// Recursively convert nested structs.
				res[name] = structToMap(value)
			}
		default:
			// For all other types (e.g., string, int, etc.), just assign the value.
			res[name] = value.Interface()
		}
	}

	return res
}
