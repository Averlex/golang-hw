package types

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid" //nolint:depguard,nolintlint
)

// EventData contains the data of the event whithout its ID.
// Pointer fields are optional.
type EventData struct {
	Title       string
	Datetime    time.Time
	Duration    time.Duration
	Description *string
	UserID      string         `db:"user_id" json:"user_id,omitempty"`     //nolint:tagliatelle
	RemindIn    *time.Duration `db:"remind_in" json:"remind_in,omitempty"` //nolint:tagliatelle
}

// Event contains the data of the event with its ID.
type Event struct {
	ID uuid.UUID `db:"id" json:"id"`
	EventData
}

// NewEventData creates a new instance of EventData with the provided parameters.
//
// It validates that the title, datetime, duration, and userID are not empty.
// If any of these fields are empty, it returns an ErrEmptyField error.
//
// It also checks that the duration is not negative, returning ErrNegativeDuration if it is.
// If the remindIn duration is negative, it returns ErrNegativeRemind.
//
// The description and remindIn fields are optional and will be set only if provided.
//
// Returns a pointer to the created EventData and nil if successful, or nil and an error if validation fails.
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

// NewEvent creates a new instance of Event with the provided parameters.
//
// Runs NewEventData to validate the event data under the hood.
//
// Returns a pointer to the created Event and nil if successful, or nil and an error if validation fails.
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

// UpdateEvent creates a new instance of Event with the given ID and the provided data.
//
// It validates that the data is not nil and returns an error if it is.
//
// Returns a pointer to the created Event and nil if successful, or nil and an error if validation fails.
func UpdateEvent(id uuid.UUID, data *EventData) (*Event, error) {
	if data == nil {
		return nil, errors.New("no data passed to update event")
	}

	return &Event{
		ID:        id,
		EventData: *data,
	}, nil
}
