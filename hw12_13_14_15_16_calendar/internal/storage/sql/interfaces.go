//go:generate mockery --name=DB --dir=. --output=mocks --filename=mock_db.go --with-expecter
//go:generate mockery --name=Tx --dir=. --output=mocks --filename=mock_tx.go --with-expecter

package sql

import (
	"context"
	"database/sql"
	"fmt"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/jmoiron/sqlx"                                                         //nolint:depguard,nolintlint
)

// DB defines the methods of sqlx.DB used in storage/sql.
type DB interface {
	ConnectContext(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error)
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close()
}

// Tx defines the methods of sqlx.Tx used in storage/sql.
type Tx interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
}

// SQLXWrapper wraps sqlx.DB to implement the DB interface for production use.
//
//nolint:revive
type SQLXWrapper struct {
	db *sqlx.DB
}

// ConnectContext implements DB.ConnectContext.
func (w *SQLXWrapper) ConnectContext(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, driverName, dataSourceName)
	if err == nil {
		w.db = db
	}
	return db, err
}

// BeginTxx implements DB.BeginTxx.
func (w *SQLXWrapper) BeginTxx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	if w.db == nil {
		return nil, fmt.Errorf("begin transaction: %w", projectErrors.ErrStorageUninitialized)
	}
	return w.db.BeginTxx(ctx, opts)
}

// Close implements DB.Close.
func (w *SQLXWrapper) Close() {
	if w.db != nil {
		w.db.Close()
	}
}

// SetDB implements DB.SetDB for testing purposes.
func (w *SQLXWrapper) SetDB(db *sqlx.DB) {
	w.db = db
}
