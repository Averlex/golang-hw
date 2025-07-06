package scheduler

import "time"

// expectedFields is a map of expected configuration fields and their default values.
var expectedFields = map[string]any{
	"retries":          int(0),
	"retry_timeout":    time.Duration(0),
	"queue_interval":   time.Duration(0),
	"cleanup_interval": time.Duration(0),
}
