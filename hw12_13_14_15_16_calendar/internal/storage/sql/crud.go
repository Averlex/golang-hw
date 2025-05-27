package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_calendar/internal/storage" //nolint:depguard,nolintlint
)

// CreateEvent creates a new event in the database. Method uses context with timeout set for Storage.
//
// All event data is validated upon call. If it fails, it returns an error.
//
// If the query is successful but the given ID is already present in the DB,
// it returns ErrDataExists.
//
// Method uses transaction to ensure the atomicity of the operation over DB.
func (s *Storage) CreateEvent(ctx context.Context, title string, datetime time.Time, duration time.Duration,
	description string, userID string, remindIn time.Duration,
) (*storage.Event, error) {
	eventData, err := storage.NewEventData(title, datetime, duration, description, userID, remindIn)
	if err != nil {
		return nil, fmt.Errorf("create new event: %w", err)
	}

	event, err := storage.NewEvent(title, eventData)
	if err != nil {
		return nil, fmt.Errorf("create new event: %w", err)
	}

	err = s.withTimeout(ctx, func(localCtx context.Context) error {
		tx, err := s.db.BeginTxx(localCtx, nil)
		if err != nil {
			return fmt.Errorf("transaction begin: %w", err)
		}

		var rollbackErr error
		defer func() {
			if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
				rollbackErr = fmt.Errorf("transaction rollback: %w", err)
			}
		}()

		query := `
		INSERT INTO events (title, datetime, duration, description, user_id, remind_in)
		VALUES (:title, :datetime, :duration, :description, :user_id, :remind_in)
		ON CONFLICT (id) DO NOTHING
		`
		res, err := tx.NamedExecContext(localCtx, query, event)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrQeuryError, err)
		}

		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("%w: %w", ErrQeuryError, err)
		}
		// Transaction successful but given ID is already present in DB.
		if n == 0 {
			return ErrDataExists
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("transaction commit: %w", err)
		}
		return rollbackErr
	})
	return event, err
}

// func UpdateEvent(ctx context.Context, title string, datetime time.Time, duration time.Duration,
// 	description string, userID string, remindIn time.Duration,
// ) error {
