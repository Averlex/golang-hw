package sql_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	tPkg "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"  //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql/mocks" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/types"     //nolint:depguard,nolintlint
	"github.com/jmoiron/sqlx"                                                           //nolint:depguard,nolintlint
	"github.com/stretchr/testify/mock"                                                  //nolint:depguard,nolintlint
	"github.com/stretchr/testify/require"                                               //nolint:depguard,nolintlint
	"github.com/stretchr/testify/suite"                                                 //nolint:depguard,nolintlint
)

// ResultMock is a mock sql.Result for RowsAffected.
type ResultMock struct {
	rowsAffected int64
}

func (r ResultMock) LastInsertId() (int64, error) {
	return 0, nil
}

func (r ResultMock) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

type StorageSuite struct {
	suite.Suite
	dbMock  *mocks.DB
	txMock  *mocks.Tx
	storage *tPkg.Storage
	ctx     context.Context
}

func (s *StorageSuite) SetupTest() {
	s.dbMock = mocks.NewDB(s.T())
	s.txMock = mocks.NewTx(s.T())
	var err error
	s.storage, err = tPkg.NewStorage(time.Second, "postgres", "host", "port", "user", "pass", "dbname",
		tPkg.WithDB(s.dbMock))
	require.NoError(s.T(), err, "expected no error creating storage")
	s.ctx = context.Background()
}

func (s *StorageSuite) TearDownTest() {
	s.dbMock.AssertExpectations(s.T())
	s.txMock.AssertExpectations(s.T())
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}

func (s *StorageSuite) TestNewStorage() {
	testCases := []struct {
		name     string
		driver   string
		expected error
	}{
		{
			name:     "valid postgres driver",
			driver:   "postgres",
			expected: nil,
		},
		{
			name:     "valid postgresql driver",
			driver:   "postgresql",
			expected: nil,
		},
		{
			name:     "unsupported driver",
			driver:   "mysql",
			expected: types.ErrUnsupportedDriver,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			storage, err := tPkg.NewStorage(time.Second, tC.driver, "host", "port", "user", "pass", "dbname")
			if tC.expected != nil {
				s.Require().Error(err, "expected error, got nil")
				s.Require().ErrorIs(err, tC.expected, "expected error does not match")
				s.Require().Nil(storage, "expected nil storage, got non-nil")
				return
			}
			s.Require().NoError(err, "expected nil, got error")
			s.Require().NotNil(storage, "expected non-nil storage, got nil")
		})
	}
}

func (s *StorageSuite) TestConnect() {
	testCases := []struct {
		name     string
		dbMockFn func()
		expected error
	}{
		{
			name: "successful connection",
			dbMockFn: func() {
				s.dbMock.On("ConnectContext", mock.Anything, "postgres", mock.Anything).Return(&sqlx.DB{}, nil).Once()
			},
			expected: nil,
		},
		{
			name: "connection timeout",
			dbMockFn: func() {
				s.storage, _ = tPkg.NewStorage(time.Nanosecond, "postgres", "host", "port", "user", "pass", "dbname",
					tPkg.WithDB(s.dbMock))
				s.dbMock.On("ConnectContext", mock.Anything, "postgres", mock.Anything).Return(nil, context.DeadlineExceeded).Once()
			},
			expected: types.ErrTimeoutExceeded,
		},
		{
			name: "uninitialized db",
			dbMockFn: func() {
				s.storage, _ = tPkg.NewStorage(time.Second, "postgres", "host", "port", "user", "pass", "dbname",
					tPkg.WithDB(nil))
				s.dbMock.On("ConnectContext", mock.Anything, "postgres",
					mock.Anything).Return(nil, types.ErrDBuninitialized).Maybe()
			},
			expected: types.ErrDBuninitialized,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbMockFn()
			err := s.storage.Connect(s.ctx)
			if tC.expected != nil {
				s.Require().Error(err, "expected error, got nil")
				s.Require().ErrorIs(err, tC.expected, "expected error does not match")
			} else {
				s.Require().NoError(err, "expected nil, got error")
			}
		})
	}
}

func (s *StorageSuite) TestClose() {
	s.Run("close database", func() {
		s.dbMock.On("Close").Return().Once()
		s.storage.Close(s.ctx)
	})
}

func (s *StorageSuite) TestCreateEvent() {
	description := "Description"
	remindIn := time.Hour
	event, _ := types.NewEvent("Test Event", time.Now(), time.Hour, description, "user1", remindIn)
	notExists := fmt.Errorf("%w", sql.ErrNoRows)

	testCases := []struct {
		name     string
		event    *types.Event
		dbMockFn func()
		txMockFn func()
		expected error
	}{
		{
			name:  "valid create",
			event: event,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(notExists).Once()
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(nil).Once()
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					*event).Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.txMock.On("Commit").Return(nil).Once()
			},
			expected: nil,
		},
		{
			name:     "nil event",
			event:    nil,
			dbMockFn: func() {}, // No DB calls expected.
			txMockFn: func() {}, // No Tx calls expected.
			expected: types.ErrNoData,
		},
		{
			name:  "event exists",
			event: event,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(nil).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: types.ErrDataExists,
		},
		{
			name:  "date busy",
			event: event,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(notExists).Once()
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Run(func(args mock.Arguments) {
					// Simulating query returning true for overlapping dates.
					dest := args.Get(1).(*bool)
					*dest = true
				}).Return(nil).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: types.ErrDateBusy,
		},
		{
			name:  "query error",
			event: event,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(notExists).Once()
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(nil).Once()
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					*event).Return(ResultMock{rowsAffected: 0}, errors.New("unknown error")).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: types.ErrQeuryError,
		},
		{
			name:  "commit error",
			event: event,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(notExists).Once()
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(nil).Once()
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					*event).Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.txMock.On("Commit").Return(errors.New("unknown error")).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: errors.New("commit error"),
		},
		{
			name:  "rollback error",
			event: event,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(nil).Once()
				s.txMock.On("Rollback").Return(errors.New("unknown error")).Once()
			},
			expected: errors.New("rollback error"),
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbMockFn()
			tC.txMockFn()
			result, err := s.storage.CreateEvent(s.ctx, tC.event)
			if tC.expected != nil {
				s.Require().Error(err, "expected error, got nil")
				if tC.name != "rollback error" && tC.name != "commit error" {
					s.Require().ErrorIs(err, tC.expected, "expected error does not match")
				}
				s.Require().Nil(result, "expected nil result, got non-nil")
			} else {
				s.Require().NoError(err, "expected nil, got error")
				s.Require().Equal(tC.event, result, "event does not match the expected one")
			}
		})
	}
}
