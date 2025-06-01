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
	"github.com/google/uuid"                                                            //nolint:depguard,nolintlint
	"github.com/jmoiron/sqlx"                                                           //nolint:depguard,nolintlint
	"github.com/stretchr/testify/mock"                                                  //nolint:depguard,nolintlint
	"github.com/stretchr/testify/require"                                               //nolint:depguard,nolintlint
	"github.com/stretchr/testify/suite"                                                 //nolint:depguard,nolintlint
)

// Common test cases names.
const (
	rollbackErrCase = "rollback error"
	commitErrCase   = "commit error"
)

// Common errors used in tests.
var (
	errNotExists  = fmt.Errorf("%w", sql.ErrNoRows) // Event was not found in the database.
	errUnknownErr = errors.New("unknown error")     // Stub error for any unscecified error cases.
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

// mockEventExists is a helper function to mock the retrieval of an existing event.
func (s *StorageSuite) mockEventExists(event *types.Event) {
	s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(1).(*types.Event)
			*dest = *event // Simulate fetching the existing event.
		}).Return(nil).Once()
}

func (s *StorageSuite) mockEventNotExists() {
	s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errNotExists).Once()
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
				s.mockEventNotExists()
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
				s.mockEventNotExists()
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
				s.mockEventNotExists()
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(nil).Once()
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					*event).Return(ResultMock{rowsAffected: 0}, errUnknownErr).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: types.ErrQeuryError,
		},
		{
			name:  commitErrCase,
			event: event,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.mockEventNotExists()
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(nil).Once()
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					*event).Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.txMock.On("Commit").Return(errUnknownErr).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: errUnknownErr,
		},
		{
			name:  rollbackErrCase,
			event: event,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(nil).Once()
				s.txMock.On("Rollback").Return(errUnknownErr).Once()
			},
			expected: errors.New(rollbackErrCase),
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbMockFn()
			tC.txMockFn()
			result, err := s.storage.CreateEvent(s.ctx, tC.event)
			if tC.expected != nil {
				s.Require().Error(err, "expected error, got nil")
				if tC.name != rollbackErrCase && tC.name != commitErrCase {
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

func (s *StorageSuite) TestUpdateEvent() {
	description := "Updated Description"
	remindIn := time.Hour
	event, _ := types.NewEvent("Test Event", time.Now(), time.Hour, description, "user1", remindIn)
	dataToUpdate, _ := types.NewEventData("Updated Event", time.Now().Add(time.Hour), 2*time.Hour,
		description, "user1", remindIn)
	eventWrongUser, _ := types.NewEvent("Test Event", time.Now(), time.Hour, description, "user2", remindIn)

	testCases := []struct {
		name     string
		id       uuid.UUID
		data     *types.EventData
		dbMockFn func()
		txMockFn func()
		expected error
	}{
		{
			name: "valid update",
			data: dataToUpdate,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					*dataToUpdate).Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.txMock.On("Commit").Return(nil).Once()
			},
			expected: nil,
		},
		{
			name:     "nil data",
			data:     nil,
			dbMockFn: func() {}, // No DB calls expected.
			txMockFn: func() {}, // No Tx calls expected.
			expected: types.ErrNoData,
		},
		{
			name: "event not found",
			id:   uuid.New(),
			data: dataToUpdate,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.mockEventNotExists()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: types.ErrEventNotFound,
		},
		{
			name: "permission denied",
			id:   event.ID,
			data: dataToUpdate,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				// Simulate an event with the same ID but different user.
				eventWrongUser.ID = event.ID
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*types.Event)
					*dest = *eventWrongUser // Simulate fetching the existing event with a different user.
				}).Return(nil).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: types.ErrPermissionDenied,
		},
		{
			name: "date busy",
			id:   event.ID,
			data: dataToUpdate,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Run(func(args mock.Arguments) {
					dest := args.Get(1).(*bool)
					*dest = true
				}).Return(nil).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: types.ErrDateBusy,
		},
		{
			name: "query error",
			id:   event.ID,
			data: dataToUpdate,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					*dataToUpdate).Return(ResultMock{rowsAffected: 1}, errUnknownErr).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: types.ErrQeuryError,
		},
		{
			name: commitErrCase,
			id:   event.ID,
			data: dataToUpdate,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					*dataToUpdate).Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.txMock.On("Commit").Return(errUnknownErr).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: errUnknownErr,
		},
		{
			name: rollbackErrCase,
			id:   event.ID,
			data: dataToUpdate,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(errUnknownErr).Once()
				s.txMock.On("Rollback").Return(errUnknownErr).Once()
			},
			expected: errors.New(rollbackErrCase),
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbMockFn()
			tC.txMockFn()
			result, err := s.storage.UpdateEvent(s.ctx, tC.id, tC.data)
			if tC.expected != nil {
				s.Require().Error(err, "expected error, got nil")
				if tC.name != rollbackErrCase && tC.name != commitErrCase {
					s.Require().ErrorIs(err, tC.expected, "expected error does not match")
				}
				s.Require().Nil(result, "expected nil result, got non-nil")
			} else {
				s.Require().NoError(err, "expected nil, got error")
				s.Require().NotNil(result, "expected non-nil result, got nil")
				s.Require().Equal(tC.id, result.ID, "event ID does not match")
			}
		})
	}
}

func (s *StorageSuite) TestDeleteEvent() {
	description := "Description"
	remindIn := time.Hour
	event, _ := types.NewEvent("Test Event", time.Now(), time.Hour, description, "user1", remindIn)

	testCases := []struct {
		name     string
		id       uuid.UUID
		dbMockFn func()
		txMockFn func()
		expected error
	}{
		{
			name: "valid delete",
			id:   event.ID,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					mock.Anything).Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.txMock.On("Commit").Return(nil).Once()
			},
			expected: nil,
		},
		{
			name: "event not found",
			id:   uuid.New(),
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.mockEventNotExists()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: types.ErrEventNotFound,
		},
		{
			name: "query error",
			id:   event.ID,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					mock.Anything).Return(ResultMock{rowsAffected: 0}, errUnknownErr).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: types.ErrQeuryError,
		},
		{
			name: commitErrCase,
			id:   event.ID,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything,
					mock.Anything).Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.txMock.On("Commit").Return(errUnknownErr).Once()
				s.txMock.On("Rollback").Return(nil).Once()
			},
			expected: errUnknownErr,
		},
		{
			name: rollbackErrCase,
			id:   event.ID,
			dbMockFn: func() {
				s.dbMock.On("BeginTxx", mock.Anything, (*sql.TxOptions)(nil)).Return(s.txMock, nil).Once()
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(errUnknownErr).Once()
				s.txMock.On("Rollback").Return(errUnknownErr).Once()
			},
			expected: errors.New(rollbackErrCase),
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbMockFn()
			tC.txMockFn()
			err := s.storage.DeleteEvent(s.ctx, tC.id)
			if tC.expected != nil {
				s.Require().Error(err, "expected error, got nil")
				if tC.name != rollbackErrCase && tC.name != commitErrCase {
					s.Require().ErrorIs(err, tC.expected, "expected error does not match")
				}
			} else {
				s.Require().NoError(err, "expected nil, got error")
			}
		})
	}
}
