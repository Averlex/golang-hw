package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require" //nolint:depguard,nolintlint
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	// Nested structure example.
	Meta struct {
		Author string `validate:"len:5"`
		Date   int    `validate:"min:2000"`
	}
	NestedUser struct {
		Name string `validate:"len:3"`
		Meta Meta   `validate:"nested"`
	}

	ComplexUser struct {
		ID     interface{}   `validate:"len:36|regexp:^\\w+$"`
		Age    interface{}   `validate:"min:18|max:50"`
		Phones []interface{} `validate:"len:11"`
	}

	InterfaceNested struct {
		Data interface{} `validate:"nested"`
	}

	InvalidRegexp struct {
		Value string `validate:"regexp:[a-z"`
	}

	DuplicateTag struct {
		Value int `validate:"min:10|min:20"`
	}

	PartialSlice struct {
		Phones []string `validate:"len:11"`
	}

	PartialNested struct {
		Name string `validate:"len:3"`
		Meta Meta   `validate:"nested"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:    "123456789012345678901234567890123456",
				Name:  "John",
				Age:   25,
				Email: "john@example.com",
				Role:  "admin",
				Phones: []string{
					"12345678901",
					"98765432109",
				},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run("common valid case", func(t *testing.T) {
			tt := tt
			t.Parallel()

			tC := tt
			err := Validate(tC.in)
			if tC.expectedErr == nil {
				require.NoError(t, err, "expected no error")
				return
			}
			var validationErrs ValidationErrors
			if errors.As(tC.expectedErr, &validationErrs) {
				var gotErrs ValidationErrors
				require.ErrorAs(t, err, &gotErrs, "expected ValidationErrors")
				require.Len(t, gotErrs, len(validationErrs), "unexpected number of validation errors")
				for _, expected := range validationErrs {
					found := false
					for _, got := range gotErrs {
						if got.Field == expected.Field {
							found = true
							break
						}
					}
					require.True(t, found, fmt.Sprintf("validation error for field %s not found", expected.Field))
				}
			} else {
				require.ErrorIs(t, err, tC.expectedErr, "unexpected error type")
			}

			_ = tt
		})
	}
}

func TestValidate_Success(t *testing.T) {
	type ManyInValues struct {
		Value int `validate:"in:200,404,500,600"`
	}

	testCases := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name: "valid app",
			in: App{
				Version: "1.0.0",
			},
			expectedErr: nil,
		},
		{
			name: "valid response",
			in: Response{
				Code: 200,
				Body: "success",
			},
			expectedErr: nil,
		},
		{
			name: "valid nested user",
			in: NestedUser{
				Name: "Bob",
				Meta: Meta{
					Author: "Alice",
					Date:   2023,
				},
			},
			expectedErr: nil,
		},
		{
			name: "valid combined tags",
			in: User{
				ID:    "123456789012345678901234567890123456",
				Age:   25,
				Email: "john@example.com",
				Role:  "admin",
				Phones: []string{
					"12345678901",
				},
			},
			expectedErr: nil,
		},
		{
			name: "valid interface fields",
			in: ComplexUser{
				ID:  "123456789012345678901234567890123456",
				Age: 25,
				Phones: []interface{}{
					"12345678901",
					"98765432109",
				},
			},
			expectedErr: nil,
		},
		{
			name: "empty slice",
			in: User{
				ID:     "123456789012345678901234567890123456",
				Age:    25,
				Email:  "john@example.com",
				Role:   "admin",
				Phones: []string{},
			},
			expectedErr: nil,
		},
		{
			name: "interface with nested struct",
			in: InterfaceNested{
				Data: NestedUser{
					Name: "Bob",
					Meta: Meta{
						Author: "Alice",
						Date:   2023,
					},
				},
			},
			expectedErr: nil,
		},

		{
			name:        "big slice for in command",
			in:          ManyInValues{Value: 200},
			expectedErr: nil,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tC.in)
			require.NoError(t, err, "expected no error")
		})
	}
}

func TestValidate_ProgramErrors(t *testing.T) {
	type InvalidTag struct {
		Value string `validate:""`
	}
	type InvalidRule struct {
		Value string `validate:"len:abc"`
	}
	type TooManyTags struct {
		Value string `validate:"len:5|regexp:^\\w+$|in:foo,bar|min:10"`
	}
	type WrongTagType struct {
		Value int `validate:"len:5"`
	}
	type TooManyRules struct {
		Value int `validate:"len:200,404,500,600"`
	}
	type PartiallyValidTags struct {
		Value string `validate:"len:5|unknown:xyz"`
	}

	testCases := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name:        "not a struct",
			in:          "not a struct",
			expectedErr: ErrInvalidData,
		},
		{
			name:        "nil input",
			in:          nil,
			expectedErr: nil,
		},
		{
			name:        "empty tag",
			in:          InvalidTag{Value: "test"},
			expectedErr: ErrInvalidData,
		},
		{
			name:        "invalid rule",
			in:          InvalidRule{Value: "test"},
			expectedErr: ErrInvalidData,
		},
		{
			name:        "too many tags",
			in:          TooManyTags{Value: "test"},
			expectedErr: ErrInvalidData,
		},
		{
			name:        "wrong tag type",
			in:          WrongTagType{Value: 42},
			expectedErr: ErrInvalidData,
		},
		{
			name:        "partially valid tags",
			in:          PartiallyValidTags{Value: "valid"},
			expectedErr: ErrInvalidData,
		},
		{
			name:        "invalid regexp",
			in:          InvalidRegexp{Value: "test"},
			expectedErr: ErrInvalidData,
		},
		{
			name:        "duplicate tag",
			in:          DuplicateTag{Value: 15},
			expectedErr: ErrInvalidData,
		},
		{
			name:        "too many rules",
			in:          TooManyRules{Value: 42},
			expectedErr: ErrInvalidData,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tC.in)
			if tC.expectedErr == nil {
				require.NoError(t, err, "expected no error")
				return
			}
			require.ErrorIs(t, err, tC.expectedErr, "unexpected error type")
		})
	}
}

func TestValidate_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name: "invalid user",
			in: User{
				ID:    "short",
				Name:  "John",
				Age:   17,
				Email: "invalid-email",
				Role:  "guest",
				Phones: []string{
					"123",
					"987654321098",
				},
			},
			expectedErr: ValidationErrors{
				{Field: "ID"},
				{Field: "Age"},
				{Field: "Email"},
				{Field: "Role"},
				{Field: "Phones"},
				{Field: "Phones"},
			},
		},
		{
			name: "invalid app",
			in: App{
				Version: "1.0",
			},
			expectedErr: ValidationErrors{
				{Field: "Version"},
			},
		},
		{
			name: "invalid response",
			in: Response{
				Code: 403,
				Body: "error",
			},
			expectedErr: ValidationErrors{
				{Field: "Code"},
			},
		},
		{
			name: "invalid nested user",
			in: NestedUser{
				Name: "Jo",
				Meta: Meta{
					Author: "Alice",
					Date:   1999,
				},
			},
			expectedErr: ValidationErrors{
				{Field: "Name"},
				{Field: "Meta.Date"},
			},
		},
		{
			name: "multiple commands",
			in: ComplexUser{
				ID:  "short",
				Age: 60,
				Phones: []interface{}{
					"123",
				},
			},
			expectedErr: ValidationErrors{
				{Field: "ID"},
				{Field: "ID"},
				{Field: "Age"},
				{Field: "Phones"},
			},
		},
		{
			name: "partially valid slice",
			in: PartialSlice{
				Phones: []string{
					"12345678901",
					"123",
				},
			},
			expectedErr: ValidationErrors{
				{Field: "Phones"},
			},
		},
		{
			name: "partially valid nested",
			in: PartialNested{
				Name: "Bob",
				Meta: Meta{
					Author: "Alice",
					Date:   1999,
				},
			},
			expectedErr: ValidationErrors{
				{Field: "Meta.Date"},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tC.in)
			var validationErrs ValidationErrors
			if errors.As(tC.expectedErr, &validationErrs) {
				var gotErrs ValidationErrors
				require.ErrorAs(t, err, &gotErrs, "expected ValidationErrors")
				require.Len(t, gotErrs, len(validationErrs), "unexpected number of validation errors")
				fieldErrCount := make(map[string]int)
				for _, err := range validationErrs {
					fieldErrCount[err.Field]++
				}
				gotFieldErrCount := make(map[string]int)
				for _, err := range gotErrs {
					gotFieldErrCount[err.Field]++
				}
				require.Equal(t, fieldErrCount, gotFieldErrCount, "mismatch in validation error counts by field")
			} else {
				require.ErrorIs(t, err, tC.expectedErr, "unexpected error type")
			}
		})
	}
}
