// Package sender provides a calendar sender, which is responsible for
// queuing the message queue for notifications and their sending.
package sender

import (
	"context"
	"fmt"
	"sync"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
)

// Sender is responsible for queuing the message queue for notifications and their sending.
type Sender struct {
	mu            sync.RWMutex
	wg            sync.WaitGroup
	l             Logger
	broker        MessageBroker
	queueInterval time.Duration
}

// NewSender creates a new calendar sender after arguments validation.
func NewSender(logger Logger, messageBrocker MessageBroker, config map[string]any) (*Sender, error) {
	// Args validation.
	missing := make([]string, 0)
	if logger == nil {
		missing = append(missing, "logger")
	}
	if messageBrocker == nil {
		missing = append(missing, "message_broker")
	}
	if config == nil {
		missing = append(missing, "config")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: some of the required parameters are missing: args=%v",
			projectErrors.ErrAppInitFailed, missing)
	}

	// Field types validation.
	missing, wrongType := validateFields(config, expectedFields)
	if len(missing) > 0 || len(wrongType) > 0 {
		return nil, fmt.Errorf("%w: missing=%v invalid_type=%v",
			projectErrors.ErrCorruptedConfig, missing, wrongType)
	}

	// Additional config vvalidation.
	invalidValues := make([]string, 0)
	queueInterval, _ := config["queue_interval"].(time.Duration)

	if queueInterval <= 0 {
		invalidValues = append(invalidValues, "queue_interval")
	}
	if len(invalidValues) > 0 {
		return nil, fmt.Errorf("%w: invalid timeout values: %v", projectErrors.ErrCorruptedConfig, invalidValues)
	}

	return &Sender{
		l:             logger,
		broker:        messageBrocker,
		queueInterval: queueInterval,
	}, nil
}

// Wait waits for the sender goroutines to finish.
func (s *Sender) Wait(_ context.Context) {
	s.wg.Wait()
}
