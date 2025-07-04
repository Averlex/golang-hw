package app

import "time"

// expectedFields is a map of expected configuration fields and their default values.
var expectedFields = map[string]any{
	"retries":       int(0),
	"retry_timeout": time.Duration(0),
}
