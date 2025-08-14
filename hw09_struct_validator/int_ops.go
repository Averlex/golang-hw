package hw09structvalidator

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// validateInts validates a slice of integers based on the provided tag.
func validateInts(vals []int, fieldName, tag string) error {
	errs := make(ValidationErrors, 0)

	cmd, err := validateCommands(tag, intCommands)
	if err != nil {
		return fmt.Errorf("%w: field=%q", err, fieldName)
	}

	for cmdName, cmdValue := range cmd {
		switch cmdName {
		case minCmd, maxCmd:
			err := validateIntBound(fieldName, cmdName, cmdValue)
			if err != nil {
				return err
			}
			errs = append(errs, checkIntBound(vals, fieldName, cmdName, cmdValue)...)
		case inCmd:
			err := validateIntIn(fieldName, cmdValue)
			if err != nil {
				return err
			}
			errs = append(errs, checkIntIn(vals, fieldName, cmdValue)...)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// validateIntBound validates the min/max instruction for integers.
func validateIntBound(fieldName, cmdName, instruction string) error {
	instructions := strings.Split(instruction, ",")
	if len(instructions) != 1 {
		return fmt.Errorf("%w: expected 1 instruction for tag %q, got %d: field=%q",
			ErrInvalidData, cmdName, len(instructions), fieldName)
	}
	if _, err := strconv.Atoi(instructions[0]); err != nil {
		return fmt.Errorf("%w: expected integer value for tag %q, got %s: field=%q",
			ErrInvalidData, cmdName, instructions[0], fieldName)
	}
	return nil
}

// checkIntBound checks integers against min/max bounds.
func checkIntBound(vals []int, fieldName, cmdName, instruction string) ValidationErrors {
	errs := make(ValidationErrors, 0)
	val, _ := strconv.Atoi(strings.Split(instruction, ",")[0])
	for i, d := range vals {
		var expectedSign string
		if cmdName == minCmd && d < val {
			expectedSign = ">="
		} else if cmdName == maxCmd && d > val {
			expectedSign = "<="
		}
		if expectedSign != "" {
			errs = errs.Add(
				fieldName,
				fmt.Errorf("int #%d value violates the %s requirement: expected %s%d, got %d", i, cmdName, expectedSign, val, d))
		}
	}
	return errs
}

// validateIntIn validates the 'in' instruction for integers.
func validateIntIn(fieldName, instruction string) error {
	instructions := strings.Split(instruction, ",")
	if len(instructions) < 1 {
		return fmt.Errorf("%w: expected >=1 instructions for tag %q, got %d: field=%q",
			ErrInvalidData, inCmd, len(instructions), fieldName)
	}
	for i, s := range instructions {
		if _, err := strconv.Atoi(s); err != nil {
			return fmt.Errorf("%w: expected integer value for #%d tag %q, got %s: field=%q",
				ErrInvalidData, i, inCmd, s, fieldName)
		}
	}
	return nil
}

// checkIntIn checks if integers are in the specified set.
func checkIntIn(vals []int, fieldName, instruction string) ValidationErrors {
	errs := make(ValidationErrors, 0)
	instructions := strings.Split(instruction, ",")
	intInstructions := make([]int, len(instructions))
	for i, s := range instructions {
		intInstructions[i], _ = strconv.Atoi(s)
	}
	for i, d := range vals {
		if !slices.Contains(intInstructions, d) {
			errs = errs.Add(
				fieldName,
				fmt.Errorf("int #%d doesn't meet the slice requirement: expected %s, got %d", i, instruction, d))
		}
	}
	return errs
}
