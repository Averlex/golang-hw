package sender

import "time"

// expectedFields is a map of expected configuration fields and their default values.
var expectedFields = map[string]any{
	"queue_interval": time.Duration(0),
}
