package hw02unpackstring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		{input: "🙃0", expected: ""},
		{input: "aaф0b", expected: "aab"},

		// Uncomment if task with asterisk completed.
		{input: `qwe\4\5`, expected: `qwe45`},
		{input: `qwe\45`, expected: `qwe44444`},
		{input: `qwe\\5`, expected: `qwe\\\\\`},
		{input: `qwe\\\3`, expected: `qwe\3`},

		// Additional test cases.
		{input: "\n", expected: "\n"},
		{input: "\n3", expected: "\n\n\n"},
		{input: "a", expected: "a"},
		{input: "a0", expected: ""},
		{input: `\\`, expected: `\`},
		{input: `\\a3`, expected: `\aaa`},
		{input: `\\\\\\`, expected: `\\\`},
		{input: `\1`, expected: `1`},
		{input: `\\0`, expected: ``},
		{input: `\\2\30`, expected: `\\`},
		{input: `a\1a`, expected: `a1a`},
		{input: `a0\\0\00`, expected: ``},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	// Added test cases. Source slice: []string{"3abc", "45", "aaa10b"}.
	invalidStrings := []string{"3abc", "45", "aaa10b", `\`, `\a`, `\\\`, `\\\a`, `ab\aba`}
	for _, tc := range invalidStrings {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}
