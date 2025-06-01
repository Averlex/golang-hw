package sql

import (
	"context"
	"fmt"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/types" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                        //nolint:depguard,nolintlint
)

// SQL queries for basic CRUD operations on events.
const (
	queryCreateEvent = `
	INSERT INTO events (title, datetime, duration, description, user_id, remind_in)
	VALUES (:title, :datetime, :duration, :description, :user_id, :remind_in)
	`
	queryUpdateEvent = `
	UPDATE events
	SET title = :title, datetime = :datetime, duration = :duration, 
	description = :description, user_id = :user_id, remind_in = :remind_in
	WHERE id = :id
	`
	queryDeleteEvent = "DELETE FROM events WHERE id = :id"
)

// CreateEvent creates a new event in the database. Method uses context with timeout set for Storage.
//
// Returns a wrapped ErrNoData error if no data passed.
//
// If the query is successful but the given ID is already present in the DB,
// it returns ErrDataExists.
//
// Method uses transaction to ensure the atomicity of the operation over DB.
func (s *Storage) CreateEvent(ctx context.Context, event *types.Event) (*types.Event, error) {
	if event == nil {
		return nil, fmt.Errorf("create new event: %w", types.ErrNoData)
	}

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx Tx) error {
		// Check if given ID is already present in DB.
		existingEvent, err := s.getExistingEvent(localCtx, tx, event.ID)
		if err != nil {
			return fmt.Errorf("create event: %w", err)
		}
		if existingEvent != nil {
			return types.ErrDataExists
		}

		isOverlaps, err := s.isOverlaps(localCtx, tx, event)
		if err != nil {
			return fmt.Errorf("create event: %w", err)
		}
		if isOverlaps {
			return types.ErrDateBusy
		}

		query := queryCreateEvent
		res, err := tx.NamedExecContext(localCtx, query, *event)
		if err != nil {
			return fmt.Errorf("%w: %w", types.ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", types.ErrQeuryError, err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	return event, nil
}

// UpdateEvent updates the event with the given ID in the database. Method uses context with timeout set for Storage.
//
// Returns a wrapped ErrNoData error if no data passed.
//
// If the query is successful but the given ID is not present in the DB, it returns ErrNotExists.
//
// Method uses transaction to ensure the atomicity of the operation over DB.
func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, data *types.EventData) (*types.Event, error) {
	if data == nil {
		return nil, fmt.Errorf("update event: %w", types.ErrNoData)
	}

	event, _ := types.UpdateEvent(id, data)

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx Tx) error {
		// Ensuring the event exists.
		existingEvent, err := s.getExistingEvent(localCtx, tx, id)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}
		if existingEvent == nil {
			return types.ErrEventNotFound
		}

		// Ensuring the event doesn't belong to another user.
		if existingEvent.UserID != event.UserID {
			return types.ErrPermissionDenied
		}

		// Ensuring the event doesn't overlap with another one.
		isOverlaps, err := s.isOverlaps(localCtx, tx, event)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}
		if isOverlaps {
			return types.ErrDateBusy
		}

		query := queryUpdateEvent
		res, err := tx.NamedExecContext(localCtx, query, *data)
		if err != nil {
			return fmt.Errorf("%w: %w", types.ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", types.ErrQeuryError, err)
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
	err := s.execInTransaction(ctx, func(localCtx context.Context, tx Tx) error {
		existingEvent, err := s.getExistingEvent(localCtx, tx, id)
		if err != nil {
			return fmt.Errorf("delete event: %w", err)
		}
		if existingEvent == nil {
			return types.ErrEventNotFound
		}

		query := queryDeleteEvent
		res, err := tx.NamedExecContext(localCtx, query, id)
		if err != nil {
			return fmt.Errorf("%w: %w", types.ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", types.ErrQeuryError, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
