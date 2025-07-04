package rabbitmq

import "time"

// expectedFields is a map of expected configuration fields and their default values.
var expectedFields = map[string]any{
	"host":     "",
	"port":     "",
	"user":     "",
	"password": "",

	"timeout":       time.Duration(0),
	"retry_timeout": time.Duration(0),
	"retries":       0,

	"topic":        "",
	"durable":      false,
	"content_type": "",

	"routing_key": "",

	"auto_ack": false,
}
