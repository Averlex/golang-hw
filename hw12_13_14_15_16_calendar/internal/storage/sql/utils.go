package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                //nolint:depguard,nolintlint
	"github.com/jmoiron/sqlx"                                               //nolint:depguard,nolintlint
)

const (
	queryGetExistingEvent = "SELECT * FROM events WHERE id = :id"
	queryIsOverlaps       = `
	SELECT EXISTS (
		SELECT 1
		FROM events
		WHERE user_id = :user_id
			AND datetime < :end_time
			AND datetime + duration > :datetime
			AND id != :id
	) AS has_conflicts
	`
)

// getExistingEvent gets the event with the given ID from the database.
// Method does not depend on database driver.
//
// Returns (nil, nil) or (*types.Event, nil) if no errors occurred, (nil, error) otherwise.
func (s *Storage) getExistingEvent(ctx context.Context, tx Tx, id uuid.UUID) (*types.Event, error) {
	var dbEvent types.DBEvent
	args := struct {
		ID uuid.UUID `db:"id"`
	}{id}
	query, qArgs, err := s.rebindQuery(queryGetExistingEvent, args)
	if err != nil {
		return nil, fmt.Errorf("event existence check: %w", err)
	}
	err = tx.GetContext(ctx, &dbEvent, query, qArgs...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("event existence check: %w", err)
	}
	return dbEvent.ToEvent(), nil
}

// isOverlaps checks if the given user event overlaps with any of his existing events in the database.
func (s *Storage) isOverlaps(ctx context.Context, tx Tx, event *types.Event) (bool, error) {
	var hasConflicts bool
	args := struct {
		UserID   string    `db:"user_id"`
		Datetime time.Time `db:"datetime"`
		EndTime  time.Time `db:"end_time"`
		ID       uuid.UUID `db:"id"`
	}{event.UserID, event.Datetime, event.Datetime.Add(event.Duration), event.ID}
	// Check the interval, excluding intersections with the event itself.
	query, qArgs, err := s.rebindQuery(queryIsOverlaps, args)
	if err != nil {
		return false, fmt.Errorf("event overlap check: %w", err)
	}
	err = tx.GetContext(ctx, &hasConflicts, query, qArgs...)
	if err != nil {
		return false, fmt.Errorf("event overlap check: %w", err)
	}

	return hasConflicts, nil
}

func (s *Storage) getBindvar() int {
	s.mu.RLock()
	driver := s.driver
	s.mu.RUnlock()

	switch driver {
	//nolint:goconst,nolintlint
	case "pgx", "postgres", "postgresql":
		return sqlx.DOLLAR // $1, $2, ...
	case "mysql", "sqlite3":
		return sqlx.QUESTION // ?
	case "oci8", "ora": // Oracle.
		return sqlx.NAMED // :arg1, :arg2.
	case "mssql", "sqlserver":
		return sqlx.AT // @arg1, @arg2.
	default:
		return sqlx.QUESTION
	}
}

// rebindQuery rebinds the query to the current bindvar of the storage.
func (s *Storage) rebindQuery(q string, args any) (string, []any, error) {
	query, qArgs, err := sqlx.Named(q, args)
	if err != nil {
		return "", nil, fmt.Errorf("rebind query: %w", err)
	}
	query = sqlx.Rebind(s.getBindvar(), query)
	return query, qArgs, nil
}
