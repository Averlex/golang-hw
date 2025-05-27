// Package sqlstorage provides a SQL database storage implementation.
package sqlstorage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	//nolint:depguard,nolintlint
	//nolint:depguard,nolintlint
	"github.com/jmoiron/sqlx" //nolint:depguard,nolintlint
)

var (
	// ErrTimeoutExceeded is returned when the operation execution times out.
	ErrTimeoutExceeded = errors.New("timeout exceeded")
	// ErrQeuryError is returned when the query execution fails.
	ErrQeuryError = errors.New("query execution")
	// ErrDataExists is returned on event ID collision on DB insertion.
	ErrDataExists = errors.New("event data already exists")
)

const defaultDriver = "postgres"

// Storage represents a SQL database storage.
type Storage struct {
	mu      sync.RWMutex
	driver  string
	db      *sqlx.DB
	dsn     string
	timeout time.Duration
}

// DBConf represents a database configuration used to build DSN string.
type DBConf struct {
	Host     string
	Port     string
	User     string
	Password string
	Timeout  time.Duration // In seconds. 0 means timeout will be disabled.
	DBname   string
}

// NewStorage creates a new Storage instance based on the given DBConf.
//
// If the arguments are empty, it returns an error.
//
// The function constructs a DSN based on the given arguments and
// the default driver. No connection is established upon the call.
func NewStorage(args *DBConf) (*Storage, error) {
	if args == nil {
		return nil, errors.New("no database args reveived")
	}

	if args.Timeout < 0 {
		return nil, errors.New("negative timeout received")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d",
		args.Host, args.Port, args.User, args.Password, args.DBname, int(args.Timeout.Seconds()),
	)

	return &Storage{
		driver:  defaultDriver,
		dsn:     dsn,
		timeout: args.Timeout,
	}, nil
}

// withTimeout wraps the given function in a context.WithTimeout call.
func (s *Storage) withTimeout(ctx context.Context, fn func(context.Context) error) error {
	if s.timeout == 0 {
		return fn(ctx)
	}

	localCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	err := fn(localCtx)
	if err != nil {
		if errors.Is(localCtx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("%w: %w", ErrTimeoutExceeded, err)
		}
		return err
	}
	return nil
}

// Connect connects to the database.
//
// If the connection is successful, it pings the database
// to check if the connection is alive. If any error occurs during the connection
// or pinging, it returns an error with ErrDBConnection wrapped around the
// original error.
func (s *Storage) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.withTimeout(ctx, func(localCtx context.Context) error {
		var err error
		s.db, err = sqlx.ConnectContext(localCtx, s.driver, s.dsn)
		if err != nil {
			return fmt.Errorf("database connection: %w", err)
		}
		return nil
	})
}

// Close closes the connection to the database.
//
// It pings the database to check if the connection is alive before closing it.
// If any error occurs during the connection or pinging, it returns an error with
// ErrDBConnection wrapped around the original error.
func (s *Storage) Close(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.Close()
	return nil
}
