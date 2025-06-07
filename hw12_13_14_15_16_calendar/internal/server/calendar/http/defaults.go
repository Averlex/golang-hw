package http

import "time"

// expectedFields is a map of expected configuration fields and their default values.
var expectedFields = map[string]any{
	"host":             "",
	"port":             "",
	"shutdown_timeout": time.Duration(0),
	"read_timeout":     time.Duration(0),
	"write_timeout":    time.Duration(0),
	"idle_timeout":     time.Duration(0),
}
