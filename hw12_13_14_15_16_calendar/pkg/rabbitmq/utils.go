package rabbitmq

import (
	"errors"
	"io"
	"syscall"
)

// isRetryable returns true if the error can be retried.
// Any connection/timout error is considered retryable.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, errTimeoutExceeded)
}
