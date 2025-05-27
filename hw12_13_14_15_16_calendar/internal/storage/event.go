package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyField       = errors.New("empty event field values received")
	ErrNegativeDuration = errors.New("event duration is negative")
	ErrNegativeRemind   = errors.New("event remind duration is negative")
	ErrGenerateID       = errors.New("failed to generate new event id")
)

type EventData struct {
	Title       string
	Datetime    time.Time
	Duration    time.Duration
	Description *string        // Optional.
	UserID      string         `db:"user_id" json:"user_id,omitempty"`
	RemindIn    *time.Duration `db:"remind_in" json:"remind_in,omitempty"` // Optional.
}

type Event struct {
	ID uuid.UUID `db:"id" json:"id"`
	EventData
}

func NewEventData(title string, datetime time.Time, duration time.Duration,
	description string, userID string, remindIn time.Duration,
) (*EventData, error) {
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

	return &EventData{
		Title:       title,
		Datetime:    datetime,
		Duration:    duration,
		Description: &desc,
		UserID:      userID,
		RemindIn:    &remind,
	}, nil
}

func NewEvent(title string, datetime time.Time, duration time.Duration,
	description string, userID string, remindIn time.Duration,
) (res *Event, err error) {
	// uuid.New() panic protection.
	defer func() {
		if r := recover(); r != nil {
			res = nil
			err = fmt.Errorf("%w: %v", ErrGenerateID, r)
		}
	}()

	data, err := NewEventData(title, datetime, duration, description, userID, remindIn)
	if err != nil {
		return nil, err
	}

	id := uuid.New()

	res = &Event{
		ID:        id,
		EventData: *data,
	}
	return
}

func UpdateEvent(id uuid.UUID, data *EventData) (*Event, error) {
	if data == nil {
		return nil, errors.New("no data passed to update event")
	}

	return &Event{
		ID:        id,
		EventData: *data,
	}, nil
}
