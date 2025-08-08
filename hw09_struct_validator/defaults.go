package hw09structvalidator

const (
	lookupTag = "validate"

	nestedTag = "nested"
)

var (
	stringTags = []string{
		"len",
		"regexp",
		"in",
	}

	intTags = []string{
		"min",
		"max",
		"in",
	}
)
