package types

import (
	"fmt"
	"time"

	"github.com/google/uuid" //nolint:depguard,nolintlint
)

// timeFormat is the default time format used for marshalling/unmarshalling.
const timeFormat = "02.01.2006 15:04:05.000"

// Notification contains the data of the notification.
type Notification struct {
	ID       string `db:"id" json:"id"`
	Title    string `db:"title" json:"title"`
	UserID   string `db:"user_id" json:"user_id"` //nolint:tagliatelle
	Datetime string `db:"datetime" json:"datetime"`
}

// GetID returns the UUID of the notification and nil on success.
// Returns an error if the UUID is invalid.
func (n *Notification) GetID() (uuid.UUID, error) {
	parsedID, err := uuid.Parse(n.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid id format in notification: %w", err)
	}
	return parsedID, nil
}

// GetDatetime returns the datetime of the notification and nil on success.
// Returns an error if the datetime is invalid.
func (n *Notification) GetDatetime() (time.Time, error) {
	parsedTime, err := time.Parse(timeFormat, n.Datetime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid datetime format in notification: %w", err)
	}
	return parsedTime, nil
}
