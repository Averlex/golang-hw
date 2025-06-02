package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                          //nolint:depguard,nolintlint
)

const (
	queryGetEventsForPeriod = `
	SELECT *
	FROM events
	WHERE datetime >= :date_start AND datetime < :date_end
	%s
	ORDER BY datetime ASC
	`
	queryGetAllUserEvents = "SELECT * FROM events WHERE user_id = :user_id"
)

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
func (s *Storage) GetEventsForDay(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error) {
	dateStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dateEnd := dateStart.AddDate(0, 0, 1)

	res, err := s.GetEventsForPeriod(ctx, dateStart, dateEnd, userID)
	if err != nil {
		return nil, fmt.Errorf("get events for day: %w", err)
	}

	return res, nil
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
func (s *Storage) GetEventsForWeek(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error) {
	// Weekday considering Monday as the first day of the week.
	weekday := (int(date.Weekday()-time.Monday) + 7) % 7

	// Truncating the date to the start of the week.
	dateStart := date.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
	dateEnd := dateStart.AddDate(0, 0, 7)

	res, err := s.GetEventsForPeriod(ctx, dateStart, dateEnd, userID)
	if err != nil {
		return nil, fmt.Errorf("get events for week: %w", err)
	}

	return res, nil
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
func (s *Storage) GetEventsForMonth(ctx context.Context, date time.Time, userID *string) ([]*types.Event, error) {
	// Truncating the date to the start of the month.
	dateStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	dateEnd := dateStart.AddDate(0, 1, 0)

	res, err := s.GetEventsForPeriod(ctx, dateStart, dateEnd, userID)
	if err != nil {
		return nil, fmt.Errorf("get events for month: %w", err)
	}

	return res, nil
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
) ([]*types.Event, error) {
	var events []*types.Event
	type Params struct {
		UserID    *string   `db:"user_id"` // Optional, can be nil.
		DateStart time.Time `db:"date_start"`
		DateEnd   time.Time `db:"date_end"`
	}
	params := Params{userID, dateStart, dateEnd}
	userIDClause := ""
	if userID != nil {
		userIDClause = "AND user_id = :user_id"
	}

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx Tx) error {
		query := fmt.Sprintf(queryGetEventsForPeriod, userIDClause)
		err := tx.SelectContext(localCtx, &events, query, params)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get events for period: %w", err)
	}
	// If no events found, set the error to ErrEventNotFound.
	if len(events) == 0 {
		return nil, fmt.Errorf("get events for period: %w", projectErrors.ErrEventNotFound)
	}

	return events, nil
}

// GetEvent retrieves an event with the specified ID from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// Returns a pointer to the Event and nil on success, or nil and any error encountered during the transaction.
// If no event with the given ID is found, it returns (nil, ErrEventNotFound).
func (s *Storage) GetEvent(ctx context.Context, id uuid.UUID) (*types.Event, error) {
	var event *types.Event
	err := s.execInTransaction(ctx, func(localCtx context.Context, tx Tx) error {
		var err error
		event, err = s.getExistingEvent(localCtx, tx, id)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("get event: %w: %w", projectErrors.ErrQeuryError, err)
	}
	if event == nil {
		return nil, fmt.Errorf("get event: %w", projectErrors.ErrEventNotFound)
	}

	return event, nil
}

// GetAllUserEvents retrieves all events for a given user ID from the database.
// The method uses a transaction with a context and timeouts as configured in Storage.
//
// Returns a slice of Event pointers and nil on success, or nil and any error encountered during the transaction.
// If no events for the given user ID are found, it returns (nil, ErrEventNotFound).
func (s *Storage) GetAllUserEvents(ctx context.Context, userID string) ([]*types.Event, error) {
	var events []*types.Event
	err := s.execInTransaction(ctx, func(localCtx context.Context, tx Tx) error {
		args := map[string]any{"user_id": userID}
		query := queryGetAllUserEvents
		err := tx.SelectContext(localCtx, &events, query, args)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get all user events: %w", err)
	}
	// If no events found, set the error to ErrEventNotFound.
	if len(events) == 0 {
		return nil, fmt.Errorf("get all user events: %w", projectErrors.ErrEventNotFound)
	}

	return events, nil
}
