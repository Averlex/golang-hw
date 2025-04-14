package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestReverse is a function to test Reverse() calls.
func TestReverse(t *testing.T) {
	// Declaring test cases
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a", expected: "a"},
		{input: "Aa", expected: "aA"},
		{input: "", expected: ""},
		{input: "1a", expected: "a1"},
		{input: "Abcdefg", expected: "gfedcbA"},
		{input: "123", expected: "321"},
		{input: `qwe\4\5`, expected: `5\4\ewq`},
		{input: `qwe\45`, expected: `54\ewq`},
		{input: `qwe\\5`, expected: `5\\ewq`},
		{input: `qwe\\\3`, expected: `3\\\ewq`},
	}

	// Running tests
	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result := Reverse(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
