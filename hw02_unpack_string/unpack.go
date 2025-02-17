package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrInvalidString = errors.New("invalid string")

// checkRune is an internal helper function, processing characters sequentially.
// It stores a buffer of 2 previous characters and returns a sequence to add.
// To process the last character of a string correctly you MUST pass an additional character.
// utf8.RuneError is considered as a neutral one. Using it as a stopper guarantees stable behavior.
func checkRune() func(char rune) (string, error) {
	const defaultChar rune = utf8.RuneError
	prevChar := defaultChar
	beforePrevChar := defaultChar
	isEscaped := false // Escape flag for previous character.

	return func(char rune) (string, error) {
		res := ""

		if d, err := strconv.Atoi(string(char)); err == nil {
			switch {
			// All error cases for digit character: sequence start, following not escaped digit.
			case (prevChar == defaultChar && beforePrevChar == defaultChar) || (unicode.IsDigit(prevChar) && !isEscaped):
				return "", ErrInvalidString
			// Digit escape case.
			case prevChar == '\\' && !isEscaped:
				isEscaped = true
			// Normal flow: digit means repeating previous character.
			default:
				isEscaped = false
				res = strings.Repeat(string(prevChar), d)
			}
		} else if char == '\\' {
			switch {
			// Escaping backslash.
			case prevChar == '\\' && !isEscaped:
				isEscaped = true
			// Character that should be written: an escaped digit or just a symbol.
			case (unicode.IsDigit(prevChar) && isEscaped) || !unicode.IsDigit(prevChar):
				isEscaped = false
				res = string(prevChar)
			// Leaving potential escape for the next iteration.
			default:
				isEscaped = false
			}
		} else {
			switch {
			// False escape attempt.
			case prevChar == '\\' && !isEscaped:
				return "", ErrInvalidString
			// Muting previous character if it was a multiplier.
			case (unicode.IsDigit(prevChar) && !isEscaped) || (prevChar == defaultChar):
				isEscaped = false
			// Normal flow: recording previous character.
			default:
				isEscaped = false
				res = string(prevChar)
			}
		}

		beforePrevChar = prevChar
		prevChar = char
		return res, nil
	}
}

// Unpack performs string unpacking. Suppors escaping '\' and digit characters.
// Example:"a4bc2e" -> "aaaabcce", `qwe\\\3` -> `qwe\\\`.
func Unpack(str string) (string, error) {
	// Checking if the string is a valid UTF-8 string.
	if !utf8.ValidString(str) {
		return "", ErrInvalidString
	}

	// Checking for an empty string as a special border case.
	if str == "" {
		return "", nil
	}

	// Organazing a structure large enough to contain the whole string.
	var builder strings.Builder
	builder.Grow(len(str))

	check := checkRune()

	// Rune by rune processing
	for _, char := range str {
		checkRes, err := check(char)
		if err != nil {
			return "", ErrInvalidString
		}

		builder.WriteString(checkRes)
	}

	// Trigger the last rune processing.
	checkRes, err := check(utf8.RuneError)
	if err != nil {
		return "", ErrInvalidString
	}

	builder.WriteString(checkRes)

	return builder.String(), nil
}
