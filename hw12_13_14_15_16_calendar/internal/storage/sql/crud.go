package sql

import (
	"context"
	"fmt"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                               //nolint:depguard,nolintlint
)

// SQL queries for basic CRUD operations on events.
const (
	queryCreateEvent = `
	INSERT INTO events (id, title, datetime, duration, description, user_id, remind_in)
	VALUES (:id, :title, :datetime, :duration, :description, :user_id, :remind_in)
	`
	queryUpdateEvent = `
	UPDATE events
	SET title = :title, datetime = :datetime, duration = :duration, 
	description = :description, user_id = :user_id, remind_in = :remind_in, is_notified = :is_notified
	WHERE id = :id
	`
	queryUpdateNotifiedEvents = `
	UPDATE events
	SET is_notified = :is_notified
	WHERE id IN (:id_list)
	`
	queryDeleteEvent     = "DELETE FROM events WHERE id = :id"
	queryDeleteOldEvents = "DELETE FROM events WHERE datetime < :date"
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
		return nil, fmt.Errorf("create event: %w", projectErrors.ErrNoData)
	}

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx Tx) error {
		// Check if given ID is already present in DB.
		existingEvent, err := s.getExistingEvent(localCtx, tx, event.ID)
		if err != nil {
			return err
		}
		if existingEvent != nil {
			return projectErrors.ErrDataExists
		}

		isOverlaps, err := s.isOverlaps(localCtx, tx, event)
		if err != nil {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}
		if isOverlaps {
			return projectErrors.ErrDateBusy
		}

		query := queryCreateEvent
		res, err := tx.NamedExecContext(localCtx, query, *event.ToDBEvent())
		if err != nil {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
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
		return nil, fmt.Errorf("update event: %w", projectErrors.ErrNoData)
	}

	event, _ := types.UpdateEvent(id, data)

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx Tx) error {
		// Ensuring the event exists.
		existingEvent, err := s.getExistingEvent(localCtx, tx, id)
		if err != nil {
			return err
		}
		if existingEvent == nil {
			return projectErrors.ErrEventNotFound
		}

		// Ensuring the event doesn't belong to another user.
		if existingEvent.UserID != event.UserID {
			return projectErrors.ErrPermissionDenied
		}

		// Ensuring the event doesn't overlap with another one.
		isOverlaps, err := s.isOverlaps(localCtx, tx, event)
		if err != nil {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}
		if isOverlaps {
			return projectErrors.ErrDateBusy
		}

		query := queryUpdateEvent
		res, err := tx.NamedExecContext(localCtx, query, event.ToDBEvent())
		if err != nil {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("update event: %w", err)
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
			return err
		}
		if existingEvent == nil {
			return projectErrors.ErrEventNotFound
		}

		queryArgs := struct {
			ID uuid.UUID `db:"id"`
		}{
			ID: id,
		}
		query := queryDeleteEvent
		res, err := tx.NamedExecContext(localCtx, query, &queryArgs)
		if err != nil {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	return nil
}

// UpdateNotifiedEvents updates the events with the given IDs in the database as notified ones.
//
// If some of the IDs are not found in the DB, they will be ignored.
//
// Returns the number of updated events and nil on success, 0 and any error otherwise.
func (s *Storage) UpdateNotifiedEvents(ctx context.Context, notifiedEvents []uuid.UUID) (int64, error) {
	var updatedCount int64
	if len(notifiedEvents) == 0 {
		return 0, fmt.Errorf("update notified events: %w", projectErrors.ErrNoData)
	}

	err := s.execInTransaction(ctx, func(localCtx context.Context, tx Tx) error {
		args := struct {
			IsNotified bool `db:"is_notified"`
		}{true}
		// The following beauty is a workaround for the named placeholders and list arg.
		// sqlx doesn't support arguments of slice type. Therefore we:
		// 	- replace the named placeholder with the list of ? placeholders;
		// 	- rebind the query for the named placeholders;
		// 	- transform the slice to []any and append it to the rebinded query arguments.
		// UUID method parameter guarantees that no injections are possible for unnamed placeholders.
		query := s.replacePlaceholder(queryUpdateNotifiedEvents, ":id_list", len(notifiedEvents))
		query, queryArgs, err := s.rebindQuery(query, args)
		if err != nil {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}
		anyEvents := make([]any, len(notifiedEvents))
		for i, id := range notifiedEvents {
			anyEvents[i] = id
		}
		// Default method flow.
		res, err := tx.ExecContext(localCtx, query, append(queryArgs, anyEvents...)...)
		if err != nil {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}

		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}
		updatedCount = n

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("update notified events: %w", err)
	}

	return updatedCount, nil
}

// DeleteOldEvents deletes all events older than the given date from the database.
// Returns the number of deleted events and nil on success, 0 and any error otherwise.
func (s *Storage) DeleteOldEvents(ctx context.Context, date time.Time) (int64, error) {
	var deletedCount int64
	err := s.execInTransaction(ctx, func(localCtx context.Context, tx Tx) error {
		queryArgs := struct {
			Date time.Time `db:"date"`
		}{date}
		query := queryDeleteOldEvents
		res, err := tx.NamedExecContext(localCtx, query, &queryArgs)
		if err != nil {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}

		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("%w: %w", projectErrors.ErrQeuryError, err)
		}
		deletedCount = n

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("delete old events: %w", err)
	}

	return deletedCount, nil
}
