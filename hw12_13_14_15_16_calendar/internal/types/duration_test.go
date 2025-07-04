package types

import (
	"database/sql/driver" //nolint:depguard,nolintlint
	"testing"
	"time"

	"github.com/stretchr/testify/require" //nolint:depguard,nolintlint
)

// TestDuration_Scan tests the Scan method of the Duration type.
func TestDuration_Scan(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected Duration
		hasError bool
	}{
		{
			name:     "string format/1 year 2 days",
			input:    "1 year 2 days",
			expected: NewDuration(365*24*time.Hour + 2*24*time.Hour),
			hasError: false,
		},
		{
			name:     "string format/2 months 3 hours",
			input:    "2 months 3 hours",
			expected: NewDuration(2*30*24*time.Hour + 3*time.Hour),
			hasError: false,
		},
		{
			name:     "string format/HH:MM:SS",
			input:    "01:30:00",
			expected: NewDuration(1*time.Hour + 30*time.Minute),
			hasError: false,
		},
		{
			name:     "string format/HH:MM:SS.ffffff",
			input:    "01:30:00.123456",
			expected: NewDuration(1*time.Hour + 30*time.Minute + 123456*time.Microsecond),
			hasError: false,
		},
		{
			name:     "string format/1 second",
			input:    "1 second",
			expected: NewDuration(1 * time.Second),
			hasError: false,
		},
		{
			name:     "string format/100 milliseconds",
			input:    "100 milliseconds",
			expected: NewDuration(100 * time.Millisecond),
			hasError: false,
		},
		{
			name:     "string format/500 microseconds",
			input:    "500 microseconds",
			expected: NewDuration(500 * time.Microsecond),
			hasError: false,
		},
		{
			name:     "int64 format/3600 seconds",
			input:    int64(3600),
			expected: NewDuration(3600 * time.Second),
			hasError: false,
		},
		{
			name:     "nil value",
			input:    nil,
			expected: NewDuration(0),
			hasError: false,
		},
		{
			name:     "invalid type",
			input:    42.0,
			expected: Duration{},
			hasError: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			var d Duration
			err := d.Scan(tC.input)
			if tC.hasError {
				require.Error(t, err, "expected error for invalid input")
				return
			}
			require.NoError(t, err, "unexpected error during Scan")
			require.Equal(t, tC.expected, d, "unexpected Duration value")
		})
	}
}

// TestDuration_Value tests the Value method of the Duration type.
func TestDuration_Value(t *testing.T) {
	testCases := []struct {
		name     string
		input    Duration
		expected driver.Value
		hasError bool
	}{
		{
			name:     "non-zero duration",
			input:    NewDuration(3600 * time.Second),
			expected: "3600 seconds",
			hasError: false,
		},
		{
			name:     "zero duration",
			input:    NewDuration(0),
			expected: "0 seconds",
			hasError: false,
		},
		{
			name:     "fractional seconds",
			input:    NewDuration(123456 * time.Microsecond),
			expected: "0 seconds",
			hasError: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			value, err := tC.input.Value()
			if tC.hasError {
				require.Error(t, err, "expected error for Value")
				return
			}
			require.NoError(t, err, "unexpected error during Value")
			require.Equal(t, tC.expected, value, "unexpected driver.Value")
		})
	}
}
