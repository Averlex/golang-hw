package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                           //nolint:depguard,nolintlint
)

const (
	queryGetExistingEvent = "SELECT * FROM events WHERE id = :id"
	queryIsOverlaps       = `
	SELECT EXISTS (
		FROM events
		WHERE user_id = :user_id
			AND datetime < :end_time
			AND datetime + duration > :datetime
			AND id != COALESCE(:event_id, '00000000-0000-0000-0000-000000000000')
	) AS has_conflicts;
	`
)

// getExistingEvent gets the event with the given ID from the database.
// Method does not depend on database driver.
//
// Returns (nil, nil) or (*types.Event, nil) if no errors occurred, (nil, error) otherwise.
func (s *Storage) getExistingEvent(ctx context.Context, tx Tx, id uuid.UUID) (*types.Event, error) {
	var event types.Event
	args := map[string]any{"id": id}
	query := queryGetExistingEvent
	err := tx.GetContext(ctx, &event, query, args)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("event existence check: %w", err)
	}
	return &event, nil
}

// isOverlaps checks if the given user event overlaps with any of his existing events in the database.
func (s *Storage) isOverlaps(ctx context.Context, tx Tx, event *types.Event) (bool, error) {
	var hasConflicts bool
	args := map[string]any{
		"user_id":  event.UserID,
		"datetime": event.Datetime,
		"end_time": event.Datetime.Add(event.Duration),
	}
	// Check the interval, excluding intersections with the event itself.
	query := queryIsOverlaps
	err := tx.GetContext(ctx, &hasConflicts, query, args)
	if err != nil {
		return false, fmt.Errorf("event overlap check: %w", err)
	}

	return hasConflicts, nil
}
