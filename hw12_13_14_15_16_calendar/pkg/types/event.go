// Package types contains Event type and its constructor and helper functions.
package types

import (
	"errors"
	"fmt"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                          //nolint:depguard,nolintlint
)

// EventData contains the data of the event whithout its ID.
// Pointer fields are optional.
type EventData struct {
	Title       string
	Datetime    time.Time
	Duration    time.Duration
	Description string
	UserID      string        `db:"user_id" json:"user_id,omitempty"`     //nolint:tagliatelle
	RemindIn    time.Duration `db:"remind_in" json:"remind_in,omitempty"` //nolint:tagliatelle
}

// DBEvent contains the data of the event with its ID.
type DBEvent struct {
	ID uuid.UUID `db:"id" json:"id"`
	DBEventData
}

// DBEventData contains the data of the event whithout its ID.
// Pointer fields are optional.
// Duration types are stored as strings.
type DBEventData struct {
	Title       string
	Datetime    time.Time
	Duration    string
	Description string
	UserID      string `db:"user_id" json:"user_id,omitempty"`     //nolint:tagliatelle
	RemindIn    string `db:"remind_in" json:"remind_in,omitempty"` //nolint:tagliatelle
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
// It also checks that the duration and remindIn are not negative, returning ErrEmptyField if they are.
//
// The description and remindIn fields are optional and will be set only if provided.
//
// Returns a pointer to the created EventData and nil if successful, or nil and an error if validation fails.
func NewEventData(title string, datetime time.Time, duration time.Duration,
	description string, userID string, remindIn time.Duration,
) (*EventData, error) {
	missing, invalid := make([]string, 0), make([]string, 0)
	if title == "" {
		missing = append(missing, "title")
	}
	if datetime.IsZero() {
		missing = append(missing, "datetime")
	}
	if duration == 0 {
		missing = append(missing, "duration")
	}
	if userID == "" {
		missing = append(missing, "userID")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: missing=%v", projectErrors.ErrEmptyField, missing)
	}

	if duration < 0 {
		invalid = append(invalid, "duration")
	}
	if remindIn < 0 {
		invalid = append(invalid, "remindIn")
	}
	if len(invalid) > 0 {
		return nil, fmt.Errorf("%w: invalid=%v", projectErrors.ErrInvalidFieldData, invalid)
	}

	return &EventData{
		Title:       title,
		Datetime:    datetime,
		Duration:    duration,
		Description: description,
		UserID:      userID,
		RemindIn:    remindIn,
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
			err = fmt.Errorf("%w: %v", projectErrors.ErrGenerateID, r)
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

// DeepCopyEvent creates a deep copy of the given Event.
func DeepCopyEvent(event *Event) *Event {
	if event == nil {
		return nil
	}

	return &Event{
		ID: event.ID,
		EventData: EventData{
			Title:       event.Title,
			Datetime:    event.Datetime,
			Duration:    event.Duration,
			Description: event.Description,
			UserID:      event.UserID,
			RemindIn:    event.RemindIn,
		},
	}
}

// ToDBEvent converts the Event to DBEvent for duration types compatibility.
func (e *Event) ToDBEvent() *DBEvent {
	return &DBEvent{
		ID:          e.ID,
		DBEventData: *e.ToDBEventData(),
	}
}

// ToDBEventData converts the EventData to DBEventData for duration types compatibility.
func (ed *EventData) ToDBEventData() *DBEventData {
	return &DBEventData{
		Title:       ed.Title,
		Datetime:    ed.Datetime,
		Duration:    fmt.Sprintf("%d", int64(ed.Duration.Seconds())),
		Description: ed.Description,
		UserID:      ed.UserID,
		RemindIn:    fmt.Sprintf("%d", int64(ed.RemindIn.Seconds())),
	}
}
