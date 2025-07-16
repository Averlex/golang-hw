// Package sender provides a calendar sender, which is responsible for
// queuing the message queue for notifications and their sending.
package sender

import (
	"context"
	"fmt"
	"sync"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
)

// Sender is responsible for queuing the message queue for notifications and their sending.
type Sender struct {
	wg     sync.WaitGroup
	l      Logger
	broker MessageBroker
}

// NewSender creates a new calendar sender after arguments validation.
func NewSender(logger Logger, messageBrocker MessageBroker) (*Sender, error) {
	// Args validation.
	missing := make([]string, 0)
	if logger == nil {
		missing = append(missing, "logger")
	}
	if messageBrocker == nil {
		missing = append(missing, "message_broker")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: some of the required parameters are missing: args=%v",
			projectErrors.ErrAppInitFailed, missing)
	}

	return &Sender{
		l:      logger,
		broker: messageBrocker,
	}, nil
}

// Wait waits for the sender goroutines to finish.
func (s *Sender) Wait(_ context.Context) {
	s.wg.Wait()
}
