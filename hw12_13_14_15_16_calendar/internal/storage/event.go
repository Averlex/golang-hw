package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	
)

var (
	ErrNoData           = errors.New("no event data received")
	ErrEmptyField       = errors.New("empty event field values received")
	ErrNegativeDuration = errors.New("event duration is negative")
	ErrNegativeRemind   = errors.New("event remind duration is negative")
	ErrGenerateID       = errors.New("failed to generate new event id")
)

type eventData struct {
	Title       string
	Datetime    time.Time
	Duration    time.Duration
	Description *string        // Optional.
	UserID      string         `db:"user_id" json:"user_id,omitempty"`
	RemindIn    *time.Duration `db:"remind_in" json:"remind_in,omitempty"` // Optional.
}

type Event struct {
	ID uuid.UUID `db:"id" json:"id"`
	eventData
}

func NewEventData(title string, datetime time.Time, duration time.Duration,
	description string, userID string, remindIn time.Duration,
) (*eventData, error) {
	var desc string
	var remind time.Duration

	if description != "" {
		desc = description
	}
	if remindIn != 0 {
		remind = remindIn
	}

	if title == "" || datetime.IsZero() || duration == 0 || userID == "" {
		return nil, ErrEmptyField
	}

	if duration < 0 {
		return nil, ErrNegativeDuration
	}

	if remindIn < 0 {
		return nil, ErrNegativeRemind
	}

	return &eventData{
		Title:       title,
		Datetime:    datetime,
		Duration:    duration,
		Description: &desc,
		UserID:      userID,
		RemindIn:    &remind,
	}, nil
}

func NewEvent(title string, data *eventData) (res *Event, err error) {
	// uuid.New() panic protection.
	defer func() {
		if r := recover(); r != nil {
			res = nil
			err = fmt.Errorf("%w: %v", ErrGenerateID, r)
		}
	}()

	if data == nil {
		return nil, ErrNoData
	}

	id := uuid.New()

	res = &Event{
		ID:        id,
		eventData: *data,
	}
	return
}
