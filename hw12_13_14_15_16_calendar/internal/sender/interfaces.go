package sender

import (
	"context"
)

// Logger represents an interface of logger visible to the app.
type Logger interface {
	// Info logs a message with level Info on the standard logger.
	Info(context.Context, string, ...any)
	// Debug logs a message with level Debug on the standard logger.
	Debug(context.Context, string, ...any)
	// Warn logs a message with level Warn on the standard logger.
	Warn(context.Context, string, ...any)
	// Error logs a message with level Error on the standard logger.
	Error(context.Context, string, ...any)
}

// MessageBroker represents a universal message broker interface.
type MessageBroker interface {
	// Consume opens a channel to receive messages from the message broker.
	// Returns data and error channels.
	Consume(context.Context) (<-chan []byte, <-chan error)
}
