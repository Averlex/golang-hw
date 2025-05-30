package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sttypes "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/storagetypes" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                                       //nolint:depguard,nolintlint
	"github.com/jmoiron/sqlx"                                                                      //nolint:depguard,nolintlint
)

// CreateEvent creates a new event in the database. Method uses context with timeout set for Storage.
//
// Returns a wrapped ErrNoData error if no data passed.
//
// If the query is successful but the given ID is already present in the DB,
// it returns ErrDataExists.
//
// Method uses transaction to ensure the atomicity of the operation over DB.
func (s *Storage) CreateEvent(ctx context.Context, event *sttypes.Event) (*sttypes.Event, error) {
	if event == nil {
		return nil, fmt.Errorf("create new event: %w", sttypes.ErrNoData)
	}

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		// Check if given ID is already present in DB.
		existingEvent, err := s.getExistingEvent(localCtx, tx, event.ID)
		if err != nil {
			return fmt.Errorf("create event: %w", err)
		}
		if existingEvent != nil {
			return sttypes.ErrDataExists
		}

		isOverlaps, err := s.isOverlaps(localCtx, tx, event)
		if err != nil {
			return fmt.Errorf("create event: %w", err)
		}
		if isOverlaps {
			return sttypes.ErrDateBusy
		}

		query := `
		INSERT INTO events (title, datetime, duration, description, user_id, remind_in)
		VALUES (:title, :datetime, :duration, :description, :user_id, :remind_in)
		`
		res, err := tx.NamedExecContext(localCtx, query, *event)
		if err != nil {
			return fmt.Errorf("%w: %w", sttypes.ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", sttypes.ErrQeuryError, err)
		}

		return nil
	})
	return event, err
}

// UpdateEvent updates the event with the given ID in the database. Method uses context with timeout set for Storage.
//
// Returns a wrapped ErrNoData error if no data passed.
//
// If the query is successful but the given ID is not present in the DB, it returns ErrNotExists.
//
// Method uses transaction to ensure the atomicity of the operation over DB.
func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, data *sttypes.EventData) (*sttypes.Event, error) {
	if data == nil {
		return nil, fmt.Errorf("update event: %w", sttypes.ErrNoData)
	}

	event, _ := sttypes.UpdateEvent(id, data)

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		// Ensuring the event exists.
		existingEvent, err := s.getExistingEvent(localCtx, tx, id)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}
		if existingEvent == nil {
			return sttypes.ErrEventNotFound
		}

		// Ensuring the event doesn't belong to another user.
		if existingEvent.UserID == event.UserID {
			return sttypes.ErrPermissionDenied
		}

		// Ensuring the event doesn't overlap with another one.
		isOverlaps, err := s.isOverlaps(localCtx, tx, event)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}
		if isOverlaps {
			return sttypes.ErrDateBusy
		}

		query := `
		UPDATE events
		SET title = :title, datetime = :datetime, duration = :duration, 
		description = :description, user_id = :user_id, remind_in = :remind_in
		WHERE id = :id
		`
		res, err := tx.NamedExecContext(localCtx, query, *data)
		if err != nil {
			return fmt.Errorf("%w: %w", sttypes.ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", sttypes.ErrQeuryError, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return event, nil
}

// DeleteEvent deletes the event with the given ID from the database. Method uses context with timeout set for Storage.
//
// If the query is successful but the given ID is not present in the DB, it returns ErrNotExists.
//
// Method uses transaction to ensure the atomicity of the operation over DB.
func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		existingEvent, err := s.getExistingEvent(localCtx, tx, id)
		if err != nil {
			return fmt.Errorf("delete event: %w", err)
		}
		if existingEvent == nil {
			return sttypes.ErrEventNotFound
		}

		query := "DELETE FROM events WHERE id = :id"
		res, err := tx.NamedExecContext(localCtx, query, id)
		if err != nil {
			return fmt.Errorf("%w: %w", sttypes.ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", sttypes.ErrQeuryError, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetEventsForDay retrieves all events occurring on the specified date from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// It fetches events where the datetime falls within the start and end of the given date,
// ordered by datetime in ascending order.
// Method truncates any given date to the start of the day.
//
// It accepts an optional userID parameter to filter events by user ID.
//
// Returns a slice of Event pointers and nil on success. If no events are found, it returns (nil, ErrEventNotFound).
// Returns nil and any error encountered during the transaction or query execution.
func (s *Storage) GetEventsForDay(ctx context.Context, date time.Time, userID *string) ([]*sttypes.Event, error) {
	dateStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dateEnd := dateStart.AddDate(0, 0, 1)

	return s.GetEventsForPeriod(ctx, dateStart, dateEnd, userID)
}

// GetEventsForWeek retrieves all events occurring on the calendar week of the specified date from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// It fetches events where the datetime falls within the start and end of the calendar week of the given date,
// ordered by datetime in ascending order.
// Method truncates any given date to the start of the calendar week.
//
// It accepts an optional userID parameter to filter events by user ID.
//
// Returns a slice of Event pointers and nil on success. If no events are found, it returns (nil, ErrEventNotFound).
// Returns nil and any error encountered during the transaction or query execution.
func (s *Storage) GetEventsForWeek(ctx context.Context, date time.Time, userID *string) ([]*sttypes.Event, error) {
	// Weekday considering Monday as the first day of the week.
	weekday := (int(date.Weekday()-time.Monday) + 7) % 7

	// Truncating the date to the start of the week.
	dateStart := date.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
	dateEnd := dateStart.AddDate(0, 0, 7)

	return s.GetEventsForPeriod(ctx, dateStart, dateEnd, userID)
}

// GetEventsForMonth retrieves all events occurring on the calendar month of the specified date from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// It fetches events where the datetime falls within the start and end of the calendar month of the given date,
// ordered by datetime in ascending order.
// Method truncates any given date to the start of the calendar month.
//
// It accepts an optional userID parameter to filter events by user ID.
//
// Returns a slice of Event pointers and nil on success. If no events are found, it returns (nil, ErrEventNotFound).
// Returns nil and any error encountered during the transaction or query execution.
func (s *Storage) GetEventsForMonth(ctx context.Context, date time.Time, userID *string) ([]*sttypes.Event, error) {
	// Truncating the date to the start of the month.
	dateStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	dateEnd := dateStart.AddDate(0, 1, 0)

	return s.GetEventsForPeriod(ctx, dateStart, dateEnd, userID)
}

// GetEventsForPeriod retrieves all events occurring on the given period from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// It fetches events where the datetime falls within the start and end of the given period,
// ordered by datetime in ascending order.
//
// It accepts an optional userID parameter to filter events by user ID.
//
// Returns a slice of Event pointers and nil on success. If no events are found, it returns (nil, ErrEventNotFound).
// Returns nil and any error encountered during the transaction or query execution.
func (s *Storage) GetEventsForPeriod(ctx context.Context, dateStart, dateEnd time.Time,
	userID *string,
) ([]*sttypes.Event, error) {
	var events []*sttypes.Event
	type Params struct {
		UserID    *string   `db:"user_id"` // Optional, can be nil.
		DateStart time.Time `db:"date_start"`
		DateEnd   time.Time `db:"date_end"`
	}
	params := Params{userID, dateStart, dateEnd}

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		query := `
		SELECT *
		FROM events
		WHERE datetime >= :date_start AND datetime < :date_end
		`
		if userID != nil {
			query += "AND user_id = :user_id "
		}
		query += "ORDER BY datetime ASC"

		return tx.SelectContext(localCtx, &events, query, params)
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", sttypes.ErrQeuryError, err)
	}
	// If no events found, set the error to ErrEventNotFound.
	if len(events) == 0 {
		return nil, sttypes.ErrEventNotFound
	}

	return events, nil
}

// getExistingEvent gets the event with the given ID from the database.
// Method does not depend on database driver.
//
// Returns (nil, nil) or (*sttypes.Event, nil) if no errors occurred, (nil, error) otherwise.
func (s *Storage) getExistingEvent(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (*sttypes.Event, error) {
	var event sttypes.Event
	args := map[string]any{"id": id}
	query := "SELECT * FROM events WHERE id = :id"
	err := tx.GetContext(ctx, &event, query, args)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("event existence check: %w", err)
	}
	return &event, nil
}

// GetEvent retrieves an event with the specified ID from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// Returns a pointer to the Event and nil on success, or nil and any error encountered during the transaction.
// If no event with the given ID is found, it returns (nil, ErrEventNotFound).
func (s *Storage) GetEvent(ctx context.Context, id uuid.UUID) (*sttypes.Event, error) {
	var event *sttypes.Event
	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		var err error
		event, err = s.getExistingEvent(localCtx, tx, id)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	if event == nil {
		return nil, sttypes.ErrEventNotFound
	}

	return event, nil
}

// GetAllUserEvents retrieves all events for a given user ID from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// Returns a slice of Event pointers and nil on success, or nil and any error encountered during the transaction.
// If no events for the given user ID are found, it returns (nil, ErrEventNotFound).
func (s *Storage) GetAllUserEvents(ctx context.Context, userID string) ([]*sttypes.Event, error) {
	var events []*sttypes.Event
	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		args := map[string]any{"user_id": userID}
		query := "SELECT * FROM events WHERE user_id = :user_id"
		query = tx.Rebind(query)
		err := tx.SelectContext(localCtx, &events, query, args)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("%w: %w", sttypes.ErrQeuryError, err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get all user events: %w", err)
	}
	if events == nil {
		return nil, sttypes.ErrEventNotFound
	}

	return events, nil
}

// isOverlaps checks if the given user event overlaps with any of his existing events in the database.
func (s *Storage) isOverlaps(ctx context.Context, tx *sqlx.Tx, event *sttypes.Event) (bool, error) {
	var hasConflicts bool
	args := map[string]any{
		"user_id":  event.UserID,
		"datetime": event.Datetime,
		"end_time": event.Datetime.Add(event.Duration),
	}
	// Check the interval, excluding intersections with the event itself.
	query := `
	SELECT EXISTS (
		FROM events
		WHERE user_id = :user_id
			AND datetime < :end_time
			AND datetime + duration > :datetime
			AND id != COALESCE(:event_id, '00000000-0000-0000-0000-000000000000')
	) AS has_conflicts;
	`
	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		return false, fmt.Errorf("event overlap check: %w", err)
	}
	err = stmt.GetContext(ctx, &hasConflicts, args)
	if err != nil {
		return false, fmt.Errorf("event overlap check: %w", err)
	}

	return hasConflicts, nil
}
