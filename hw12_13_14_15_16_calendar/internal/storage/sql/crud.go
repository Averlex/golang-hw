package sqlstorage

import (
	"context"
	"fmt"

	"github.com/Averlex/golang-hw/hw12_13_14_15_calendar/internal/storage" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                               //nolint:depguard,nolintlint
	"github.com/jmoiron/sqlx"                                              //nolint:depguard,nolintlint
)

// CreateEvent creates a new event in the database. Method uses context with timeout set for Storage.
//
// Returns a wrapped ErrNoData error if no data passed.
//
// If the query is successful but the given ID is already present in the DB,
// it returns ErrDataExists.
//
// Method uses transaction to ensure the atomicity of the operation over DB.
func (s *Storage) CreateEvent(ctx context.Context, event *storage.Event) (*storage.Event, error) {
	if event == nil {
		return nil, fmt.Errorf("create new event: %w", ErrNoData)
	}

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		// Check if given ID is already present in DB.
		ok, err := s.isExists(localCtx, tx, event.ID)
		if err != nil {
			return fmt.Errorf("create event: %w", err)
		}
		if ok {
			return ErrDataExists
		}

		query := `
		INSERT INTO events (title, datetime, duration, description, user_id, remind_in)
		VALUES (:title, :datetime, :duration, :description, :user_id, :remind_in)
		`
		res, err := tx.NamedExecContext(localCtx, query, *event)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", ErrQeuryError, err)
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
func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, data *storage.EventData) (*storage.Event, error) {
	if data == nil {
		return nil, fmt.Errorf("update event: %w", ErrNoData)
	}

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx *sqlx.Tx) error {
		ok, err := s.isExists(localCtx, tx, id)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}
		if !ok {
			return ErrNotExists
		}

		query := `
		UPDATE events
		SET title = :title, datetime = :datetime, duration = :duration, 
		description = :description, user_id = :user_id, remind_in = :remind_in
		WHERE id = :id
		`
		res, err := tx.NamedExecContext(localCtx, query, *data)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", ErrQeuryError, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	event, _ := storage.UpdateEvent(id, data)

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
			return ErrNotExists
		}

		query := "DELETE FROM events WHERE id = :id"
		res, err := tx.NamedExecContext(localCtx, query, id)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", ErrQeuryError, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
