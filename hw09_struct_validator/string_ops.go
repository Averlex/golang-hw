package hw09structvalidator

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

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

	for cmdName, cmdValue := range cmd {
		switch cmdName {
		case lenCmd:
			err := validateStringLength(fieldName, cmdValue)
			if err != nil {
				return err
			}
			errs = append(errs, checkStringLength(vals, fieldName, cmdValue)...)
		case regexpCmd:
			err := validateStringRegexp(fieldName, cmdValue)
			if err != nil {
				return err
			}
			errs = append(errs, checkStringRegexp(vals, fieldName, cmdValue)...)
		case inCmd:
			err := validateStringIn(fieldName, cmdValue)
			if err != nil {
				return err
			}
			errs = append(errs, checkStringIn(vals, fieldName, cmdValue)...)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// validateStringLength validates the length instruction for strings.
func validateStringLength(fieldName, instruction string) error {
	instructions := strings.Split(instruction, ",")
	if len(instructions) != 1 {
		return fmt.Errorf("%w: expected 1 instruction for tag %q, got %d: field=%q",
			ErrInvalidData, lenCmd, len(instructions), fieldName)
	}
	if _, err := strconv.Atoi(instructions[0]); err != nil {
		return fmt.Errorf("%w: expected integer value for tag %q, got %s: field=%q",
			ErrInvalidData, lenCmd, instructions[0], fieldName)
	}
	return nil
}

// checkStringLength checks string lengths against the specified length.
func checkStringLength(vals []string, fieldName, instruction string) ValidationErrors {
	errs := make(ValidationErrors, 0)
	length, _ := strconv.Atoi(strings.Split(instruction, ",")[0])
	for i, s := range vals {
		if len(s) != length {
			errs = errs.Add(
				fieldName,
				fmt.Errorf("string #%d length doesn't meet the length requirement: expected %d, got %d", i, length, len(s)))
		}
	}
	return errs
}

// validateStringRegexp validates the regular expression for strings.
func validateStringRegexp(fieldName, expression string) error {
	if _, err := regexp.Compile(expression); err != nil {
		return fmt.Errorf("%w: expected valid regular expression for tag %q, got %s: field=%q",
			ErrInvalidData, regexpCmd, expression, fieldName)
	}
	return nil
}

// checkStringRegexp checks strings against the regular expression.
func checkStringRegexp(vals []string, fieldName, expression string) ValidationErrors {
	errs := make(ValidationErrors, 0)
	compiledExp, _ := regexp.Compile(expression)
	for i, s := range vals {
		if !compiledExp.MatchString(s) {
			errs = errs.Add(
				fieldName,
				fmt.Errorf("string #%d doesn't meet the regexp requirement: expected %s, got %s", i, expression, s))
		}
	}
	return errs
}

// validateStringIn validates the 'in' instruction for strings.
func validateStringIn(fieldName, instruction string) error {
	instructions := strings.Split(instruction, ",")
	if len(instructions) < 1 {
		return fmt.Errorf("%w: expected >=1 instructions for tag %q, got %d: field=%q",
			ErrInvalidData, inCmd, len(instructions), fieldName)
	}
	return nil
}

// checkStringIn checks if strings are in the specified set.
func checkStringIn(vals []string, fieldName, instruction string) ValidationErrors {
	errs := make(ValidationErrors, 0)
	instructions := strings.Split(instruction, ",")
	for i, s := range vals {
		if !slices.Contains(instructions, s) {
			errs = errs.Add(
				fieldName,
				fmt.Errorf("string #%d doesn't meet the slice requirement: expected %s, got %s", i, instruction, s))
		}
	}
	return errs
}
