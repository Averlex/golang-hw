// Package scheduler provides a calendar scheduler, which is responsible for
// queuing the storage for events which need notifications and cleaning up
// old events from the storage.
package scheduler

import (
	"fmt"
	"sync"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
)

// Scheduler is a calendar scheduler.
type Scheduler struct {
	mu              sync.RWMutex
	wg              sync.WaitGroup
	l               Logger
	s               Storage
	broker          MessageBroker
	retries         int
	retryTimeout    time.Duration
	queueInterval   time.Duration
	cleanupInterval time.Duration
}

// NewScheduler creates a new calendar application after arguments validation.
//
// It uses the provided logger and storage to log and store events.
func NewScheduler(
	logger Logger,
	storage Storage,
	messageBrocker MessageBroker,
	config map[string]any,
) (*Scheduler, error) {
	// Args validation.
	missing := make([]string, 0)
	if logger == nil {
		missing = append(missing, "logger")
	}
	if storage == nil {
		missing = append(missing, "storage")
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

	// Extract from config an normalize the value.
	retries, _ := config["retries"].(int)
	retries = max(0, retries)
	retryTimeout, _ := config["retry_timeout"].(time.Duration)
	queueInterval, _ := config["queue_interval"].(time.Duration)
	cleanupInterval, _ := config["cleanup_interval"].(time.Duration)

	// Validation.
	invalidValues := make([]string, 0)
	if retryTimeout <= 0 {
		invalidValues = append(invalidValues, "retry_timeout")
	}
	if queueInterval <= 0 {
		invalidValues = append(invalidValues, "queue_interval")
	}
	if cleanupInterval <= 0 {
		invalidValues = append(invalidValues, "cleanup_interval")
	}
	if len(invalidValues) > 0 {
		return nil, fmt.Errorf("%w: invalid timeout values: %v", projectErrors.ErrCorruptedConfig, invalidValues)
	}

	return &Scheduler{
		l:               logger,
		s:               storage,
		broker:          messageBrocker,
		retries:         retries,
		retryTimeout:    retryTimeout,
		queueInterval:   queueInterval,
		cleanupInterval: cleanupInterval,
	}, nil
}
