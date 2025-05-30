package sqlstorage

import (
	"context"
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
		ok, err := s.isExists(localCtx, tx, event.ID)
		if err != nil {
			return fmt.Errorf("create event: %w", err)
		}
		if ok {
			return sttypes.ErrDataExists
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

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		ok, err := s.isExists(localCtx, tx, id)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}
		if !ok {
			return sttypes.ErrEventNotFound
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

	event, _ := sttypes.UpdateEvent(id, data)

	return event, nil
}

// DeleteEvent deletes the event with the given ID from the database. Method uses context with timeout set for Storage.
//
// If the query is successful but the given ID is not present in the DB, it returns ErrNotExists.
//
// Method uses transaction to ensure the atomicity of the operation over DB.
func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		ok, err := s.isExists(localCtx, tx, id)
		if err != nil {
			return fmt.Errorf("delete event: %w", err)
		}
		if !ok {
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
// Returns a slice (possibly an empty one) of Event pointers and nil on success.
// Returns nil and any error encountered during the transaction or query execution.
func (s *Storage) GetEventsForDay(ctx context.Context, date time.Time) ([]*sttypes.Event, error) {
	dateStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dateEnd := dateStart.AddDate(0, 0, 1)

	return s.getEventsForPeriod(ctx, dateStart, dateEnd)
}

// GetEventsForWeek retrieves all events occurring on the calendar week of the specified date from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// It fetches events where the datetime falls within the start and end of the calendar week of the given date,
// ordered by datetime in ascending order.
// Method truncates any given date to the start of the calendar week.
//
// Returns a slice (possibly an empty one) of Event pointers and nil on success.
// Returns nil and any error encountered during the transaction or query execution.
func (s *Storage) GetEventsForWeek(ctx context.Context, date time.Time) ([]*sttypes.Event, error) {
	// Weekday considering Monday as the first day of the week.
	weekday := (int(date.Weekday()-time.Monday) + 7) % 7

	// Truncating the date to the start of the week.
	dateStart := date.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
	dateEnd := dateStart.AddDate(0, 0, 7)

	return s.getEventsForPeriod(ctx, dateStart, dateEnd)
}

// GetEventsForMonth retrieves all events occurring on the calendar month of the specified date from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// It fetches events where the datetime falls within the start and end of the calendar month of the given date,
// ordered by datetime in ascending order.
// Method truncates any given date to the start of the calendar month.
//
// Returns a slice (possibly an empty one) of Event pointers and nil on success.
// Returns nil and any error encountered during the transaction or query execution.
func (s *Storage) GetEventsForMonth(ctx context.Context, date time.Time) ([]*sttypes.Event, error) {
	// Truncating the date to the start of the month.
	dateStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	dateEnd := dateStart.AddDate(0, 1, 0)

	return s.getEventsForPeriod(ctx, dateStart, dateEnd)
}

// getEventsForPeriod retrieves all events occurring on the given period from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// It fetches events where the datetime falls within the start and end of the given period,
// ordered by datetime in ascending order.
//
// Returns a slice (possibly an empty one) of Event pointers and nil on success.
// Returns nil and any error encountered during the transaction or query execution.
func (s *Storage) getEventsForPeriod(ctx context.Context, dateStart, dateEnd time.Time) ([]*sttypes.Event, error) {
	var events []*sttypes.Event
	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		type Params struct {
			DateStart time.Time `db:"date_start"`
			DateEnd   time.Time `db:"date_end"`
		}
		params := Params{dateStart, dateEnd}

		query := `
		SELECT *
		FROM events
		WHERE datetime >= :date_start AND datetime < :date_end
		ORDER BY datetime ASC
		`

		err := tx.SelectContext(localCtx, &events, query, params)
		if err != nil {
			return fmt.Errorf("%w: %w", sttypes.ErrQeuryError, err)
		}
		return nil
	})

	// If no events found, return empty slice instead of nil.
	if len(events) == 0 && err != nil {
		events = []*sttypes.Event{}
	}
	// If error occurred, return nil istead of any possible results.
	if err != nil {
		events = nil
	}

	return events, err
}
