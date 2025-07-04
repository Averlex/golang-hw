package http

import "time"

const swaggerPath = "api/calendar/v1/CalendarService.swagger.json"

// expectedHTTPFields is a map of expected configuration fields and their default values.
var expectedHTTPFields = map[string]any{
	"host":             "",
	"port":             "",
	"shutdown_timeout": time.Duration(0),
	"read_timeout":     time.Duration(0),
	"write_timeout":    time.Duration(0),
	"idle_timeout":     time.Duration(0),
}

// expectedGRPCFields is a map of expected configuration fields and their default values for a linked gRPC server.
var expectedGRPCFields = map[string]any{
	"host": "",
	"port": "",
}
