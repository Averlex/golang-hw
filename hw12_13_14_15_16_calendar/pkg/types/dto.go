package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid" //nolint:depguard,nolintlint
)

// Period represents a period of time.
type Period uint8

// Possible values for Period.
const (
	Day Period = iota
	Week
	Month
)

// String returns a string representation of the Period.
func (p Period) String() string {
	return [...]string{"day", "week", "month"}[p]
}

// MarshalJSON impelements custom JSON marshaling for Period type.
func (p Period) MarshalJSON() ([]byte, error) {
	// Сериализуем как строку, используя String().
	return json.Marshal(p.String())
}

// CreateEventInput represents the input for creating an event.
//
//nolint:tagliatelle
type CreateEventInput struct {
	Title       string         `json:"title"`
	Datetime    time.Time      `json:"start_date"`
	Duration    time.Duration  `json:"end_date"`
	UserID      string         `json:"user_id"`
	Description *string        `json:"description,omitempty"`
	RemindIn    *time.Duration `json:"remind_in,omitempty"`
}

// UpdateEventInput represents the input for updating an event.
//
//nolint:tagliatelle
type UpdateEventInput struct {
	ID          uuid.UUID      `json:"id"`
	Title       *string        `json:"title,omitempty"`
	Datetime    *time.Time     `json:"start_date,omitempty"`
	Duration    *time.Duration `json:"end_date,omitempty"`
	UserID      *string        `json:"user_id,omitempty"`
	Description *string        `json:"description,omitempty"`
	RemindIn    *time.Duration `json:"remind_in,omitempty"`
}

// DateFilterInput represents the input for getters by a fixed period, starting from a specific date.
//
//nolint:tagliatelle
type DateFilterInput struct {
	Date   time.Time `json:"date"`
	UserID *string   `json:"user_id"`
	Period Period    `json:"period"`
}

// DateRangeInput represents the input for getters by a range of dates.
//
//nolint:tagliatelle
type DateRangeInput struct {
	DateStart time.Time `json:"date_start"`
	DateEnd   time.Time `json:"date_end"`
	UserID    *string   `json:"user_id"`
}
