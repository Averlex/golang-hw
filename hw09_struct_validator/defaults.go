package hw09structvalidator

const (
	lookupTag = "validate"

	nestedTag = "nested"

	tagLimit       = 3 // Max number of tags equals to the number of possible different tags.
	cmdPartsNumber = 2 // Number of parts for each command. For example, "len:5" has 2 parts, same as "in:1,10,20".

	lenCmd    = "len"
	regexpCmd = "regexp"
	minCmd    = "min"
	maxCmd    = "max"
	inCmd     = "in"
)

var (
	stringCommands = []string{
		lenCmd,
		regexpCmd,
		inCmd,
	}

	intCommands = []string{
		minCmd,
		maxCmd,
		inCmd,
	}
)
