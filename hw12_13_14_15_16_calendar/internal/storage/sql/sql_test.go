package sql_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
	tPkg "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"     //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql/mocks"    //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                               //nolint:depguard,nolintlint
	"github.com/jmoiron/sqlx"                                                              //nolint:depguard,nolintlint
	"github.com/stretchr/testify/mock"                                                     //nolint:depguard,nolintlint
	"github.com/stretchr/testify/require"                                                  //nolint:depguard,nolintlint
	"github.com/stretchr/testify/suite"                                                    //nolint:depguard,nolintlint
)

const (
	// Common test cases names.
	rollbackErrCase = "rollback error"
	commitErrCase   = "commit error"
	// Common test data.
	description        = "Description"
	duration, remindIn = time.Hour, time.Hour
)

// Common errors used in tests.
var (
	errNotExists  = fmt.Errorf("%w", sql.ErrNoRows) // Event was not found in the database.
	errUnknownErr = errors.New("unknown error")     // Stub error for any unscecified error cases.
)

type SQLSuite struct {
	suite.Suite
	dbMock  *mocks.DB
	txMock  *mocks.Tx
	storage *tPkg.Storage
	ctx     context.Context
}

func (s *SQLSuite) SetupTest() {
	s.dbMock = mocks.NewDB(s.T())
	s.txMock = mocks.NewTx(s.T())
	var err error
	s.storage, err = tPkg.NewStorage(time.Second, "postgres", "host", "port", "user", "pass", "dbname",
		tPkg.WithDB(s.dbMock))
	require.NoError(s.T(), err, "expected no error creating storage")
	s.ctx = context.Background()
}

func (s *SQLSuite) TearDownTest() {
	s.dbMock.AssertExpectations(s.T())
	s.txMock.AssertExpectations(s.T())
}

func TestSQLStorage(t *testing.T) {
	suite.Run(t, new(SQLSuite))
}

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

// mockEventExists is a helper function to mock the retrieval of an existing event.
func (s *SQLSuite) mockEventExists(event *types.Event) {
	s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(1).(*types.DBEvent)
			*dest = *event.ToDBEvent() // Simulate fetching the existing event.
		}).Return(nil).Once()
}

// mockEventNotExists is a helper function to mock the case when an event does not exist.
func (s *SQLSuite) mockEventNotExists() {
	s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errNotExists).Once()
}

func (s *SQLSuite) mockEventOverlaps(isOverlaps bool) {
	// We expect 7 arguments for the GetContext call:
	// 3 necessary + variadic of 4 arguments.
	callArgs := make([]any, 7)
	for i := range callArgs {
		callArgs[i] = mock.Anything
	}
	if !isOverlaps {
		s.txMock.On("GetContext", callArgs...).Return(nil).Once()
		return
	}
	s.txMock.On("GetContext", callArgs...).Run(func(args mock.Arguments) {
		// Simulating query returning true for overlapping dates.
		dest := args.Get(1).(*bool)
		*dest = true
	}).Return(nil).Once()
}

func (s *SQLSuite) mockGetEvents(events *[]*types.DBEvent, isFound bool, argLen int) {
	args := make([]any, argLen)
	for i := range args {
		args[i] = mock.Anything
	}
	callArgs := append([]any{
		mock.Anything,
		mock.Anything,
		mock.Anything,
	}, args...)
	if !isFound {
		s.txMock.On("SelectContext", callArgs...).
			Run(func(args mock.Arguments) {
				dest := args.Get(1).(*[]*types.DBEvent)
				*dest = []*types.DBEvent{} // Simulate no events found.
			}).Return(nil).Once()
		return
	}
	s.txMock.On("SelectContext", callArgs...).
		Run(func(args mock.Arguments) {
			dest := args.Get(1).(*[]*types.DBEvent)
			*dest = *events
		}).Return(nil).Once()
}

// mockBeginTx is a helper function to mock the beginning of a transaction.
func (s *SQLSuite) mockBeginTx(success bool) {
	if !success {
		s.dbMock.On("BeginTxx", mock.Anything, mock.Anything).Return(s.txMock, errUnknownErr).Once()
		return
	}
	s.dbMock.On("BeginTxx", mock.Anything, mock.Anything).Return(s.txMock, nil).Once()
}

// mockCommit is a helper function to mock the commit of a transaction.
func (s *SQLSuite) mockCommit(success bool) {
	if !success {
		s.txMock.On("Commit").Return(errUnknownErr).Once()
		return
	}
	s.txMock.On("Commit").Return(nil).Once()
}

// mockRollback is a helper function to mock the rollback of a transaction.
func (s *SQLSuite) mockRollback(success bool) {
	if !success {
		s.txMock.On("Rollback").Return(errUnknownErr).Once()
		return
	}
	s.txMock.On("Rollback").Return(nil).Once()
}

// newTestEvent creates a new test event with the given title and userID.
// Method uses the common test data for duration, description, and remindIn.
func (s *SQLSuite) newTestEvent(title, userID string) *types.Event {
	event, _ := types.NewEvent(title, time.Now(), duration, description, userID, remindIn)
	return event
}

// newTestEventData creates a new test event data with the given title and userID.
func (s *SQLSuite) newTestEventData(title, userID string) *types.EventData {
	data, _ := types.NewEventData(title, time.Now(), duration, description, userID, remindIn)
	return data
}

func (s *SQLSuite) TestNewStorage() {
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
			expected: projectErrors.ErrUnsupportedDriver,
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

func (s *SQLSuite) TestConnect() {
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
			expected: projectErrors.ErrTimeoutExceeded,
		},
		{
			name: "uninitialized db",
			dbMockFn: func() {
				s.storage, _ = tPkg.NewStorage(time.Second, "postgres", "host", "port", "user", "pass", "dbname",
					tPkg.WithDB(nil))
				s.dbMock.On("ConnectContext", mock.Anything, "postgres",
					mock.Anything).Return(nil, projectErrors.ErrStorageUninitialized).Maybe()
			},
			expected: projectErrors.ErrStorageUninitialized,
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

func (s *SQLSuite) TestClose() {
	s.Run("close database", func() {
		s.dbMock.On("Close").Return().Once()
		s.storage.Close(s.ctx)
	})
}

func (s *SQLSuite) TestCreateEvent() {
	event := s.newTestEvent("Create event", "user1")

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
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventNotExists()
				s.mockEventOverlaps(false)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).
					Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.mockCommit(true)
			},
			expected: nil,
		},
		{
			name:     "nil event",
			event:    nil,
			dbMockFn: func() {}, // No DB calls expected.
			txMockFn: func() {}, // No Tx calls expected.
			expected: projectErrors.ErrNoData,
		},
		{
			name:  "event exists",
			event: event,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.mockRollback(true)
			},
			expected: projectErrors.ErrDataExists,
		},
		{
			name:  "date busy",
			event: event,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventNotExists()
				s.mockEventOverlaps(true)
				s.mockRollback(true)
			},
			expected: projectErrors.ErrDateBusy,
		},
		{
			name:  "query error",
			event: event,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventNotExists()
				s.mockEventOverlaps(false)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).
					Return(ResultMock{rowsAffected: 0}, errUnknownErr).Once()
				s.mockRollback(true)
			},
			expected: projectErrors.ErrQeuryError,
		},
		{
			name:  "begin transaction error",
			event: event,
			dbMockFn: func() {
				s.mockBeginTx(false)
			},
			expected: errUnknownErr,
		},
		{
			name:  commitErrCase,
			event: event,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventNotExists()
				s.mockEventOverlaps(false)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).
					Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.mockCommit(false)
				s.mockRollback(true)
			},
			expected: errUnknownErr,
		},
		{
			name:  rollbackErrCase,
			event: event,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Once()
				s.mockRollback(false)
			},
			expected: errors.New(rollbackErrCase),
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbMockFn()
			if tC.name != "begin transaction error" {
				tC.txMockFn()
			}
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

func (s *SQLSuite) TestUpdateEvent() {
	event := s.newTestEvent("Update event", "user1")
	dataToUpdate := s.newTestEventData("Update event", "user1")
	updEvent := &types.Event{EventData: *dataToUpdate}
	eventWrongUser := s.newTestEvent("Update event", "user2")

	testCases := []struct {
		name     string
		id       uuid.UUID
		data     *types.Event
		dbMockFn func()
		txMockFn func()
		expected error
	}{
		{
			name: "valid update",
			data: updEvent,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.mockEventOverlaps(false)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).
					Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.mockCommit(true)
			},
			expected: nil,
		},
		{
			name:     "nil data",
			data:     nil,
			dbMockFn: func() {}, // No DB calls expected.
			txMockFn: func() {}, // No Tx calls expected.
			expected: projectErrors.ErrNoData,
		},
		{
			name: "event not found",
			id:   uuid.New(),
			data: updEvent,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventNotExists()
				s.mockRollback(true)
			},
			expected: projectErrors.ErrEventNotFound,
		},
		{
			name: "permission denied",
			id:   event.ID,
			data: updEvent,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				// Simulate an event with the same ID but different user.
				eventWrongUser.ID = event.ID
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						dest := args.Get(1).(*types.DBEvent)
						*dest = *eventWrongUser.ToDBEvent() // Simulate fetching the existing event with a different user.
					}).Return(nil).Once()
				s.mockRollback(true)
			},
			expected: projectErrors.ErrPermissionDenied,
		},
		{
			name: "date busy",
			id:   event.ID,
			data: updEvent,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.mockEventOverlaps(true)
				s.mockRollback(true)
			},
			expected: projectErrors.ErrDateBusy,
		},
		{
			name: "query error",
			id:   event.ID,
			data: updEvent,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.mockEventOverlaps(false)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).
					Return(ResultMock{rowsAffected: 1}, errUnknownErr).Once()
				s.mockRollback(true)
			},
			expected: projectErrors.ErrQeuryError,
		},
		{
			name: commitErrCase,
			id:   event.ID,
			data: updEvent,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.mockEventOverlaps(false)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).
					Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.mockCommit(false)
				s.mockRollback(true)
			},
			expected: errUnknownErr,
		},
		{
			name: rollbackErrCase,
			id:   event.ID,
			data: updEvent,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errUnknownErr).Once()
				s.mockRollback(false)
			},
			expected: errors.New(rollbackErrCase),
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbMockFn()
			tC.txMockFn()
			var data *types.EventData
			if tC.data != nil {
				data = &tC.data.EventData
			}
			result, err := s.storage.UpdateEvent(s.ctx, tC.id, data)
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

func (s *SQLSuite) TestDeleteEvent() {
	event := s.newTestEvent("Delete event", "user1")

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
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).
					Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.mockCommit(true)
			},
			expected: nil,
		},
		{
			name: "event not found",
			id:   uuid.New(),
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventNotExists()
				s.mockRollback(true)
			},
			expected: projectErrors.ErrEventNotFound,
		},
		{
			name: "query error",
			id:   event.ID,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).
					Return(ResultMock{rowsAffected: 0}, errUnknownErr).Once()
				s.mockRollback(true)
			},
			expected: projectErrors.ErrQeuryError,
		},
		{
			name: commitErrCase,
			id:   event.ID,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.txMock.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).
					Return(ResultMock{rowsAffected: 1}, nil).Once()
				s.mockCommit(false)
				s.mockRollback(true)
			},
			expected: errUnknownErr,
		},
		{
			name: rollbackErrCase,
			id:   event.ID,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errUnknownErr).Once()
				s.mockRollback(false)
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

func (s *SQLSuite) TestGetEvent() {
	event := s.newTestEvent("Get event", "user1")

	testCases := []struct {
		name     string
		id       uuid.UUID
		dbMockFn func()
		txMockFn func()
		expected error
	}{
		{
			name: "valid get",
			id:   event.ID,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventExists(event)
				s.mockCommit(true)
			},
			expected: nil,
		},
		{
			name: "event not found",
			id:   uuid.New(),
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockEventNotExists()
				s.mockCommit(true)
			},
			expected: projectErrors.ErrEventNotFound,
		},
		{
			name: "query error",
			id:   event.ID,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.txMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errUnknownErr).Once()
				s.mockRollback(true)
			},
			expected: projectErrors.ErrQeuryError,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbMockFn()
			tC.txMockFn()
			result, err := s.storage.GetEvent(s.ctx, tC.id)
			if tC.expected != nil {
				s.Require().Error(err, "expected error, got nil")
				if tC.name != rollbackErrCase && tC.name != commitErrCase {
					s.Require().ErrorIs(err, tC.expected, "expected error does not match")
				}
				s.Require().Nil(result, "expected nil result, got non-nil")
			} else {
				s.Require().NoError(err, "expected nil, got error")
				s.Require().Equal(event, result, "event does not match the expected one")
			}
		})
	}
}

func (s *SQLSuite) TestGetAllUserEvents() {
	userID := "user1"
	events := []*types.DBEvent{
		s.newTestEvent("Event 1", userID).ToDBEvent(),
		s.newTestEvent("Event 2", userID).ToDBEvent(),
	}

	testCases := []struct {
		name     string
		userID   string
		dbMockFn func()
		txMockFn func()
		expected error
	}{
		{
			name:   "valid get all",
			userID: userID,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockGetEvents(&events, true, 1)
				s.mockCommit(true)
			},
			expected: nil,
		},
		{
			name:     "no events",
			userID:   userID,
			dbMockFn: func() { s.mockBeginTx(true) },
			txMockFn: func() {
				s.mockGetEvents(&events, false, 1)
				s.mockCommit(true)
			},
			expected: projectErrors.ErrEventNotFound,
		},
		{
			name:     "query error",
			userID:   userID,
			dbMockFn: func() { s.mockBeginTx(true) },
			txMockFn: func() {
				s.txMock.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errUnknownErr).Once()
				s.mockRollback(true)
			},
			expected: projectErrors.ErrQeuryError,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbMockFn()
			tC.txMockFn()
			result, err := s.storage.GetAllUserEvents(s.ctx, tC.userID)
			if tC.expected != nil {
				s.Require().Error(err, "expected error, got nil")
				s.Require().ErrorIs(err, tC.expected, "expected error does not match")
				s.Require().Nil(result, "expected nil result, got non-nil")
			} else {
				s.Require().NoError(err, "expected nil, got error")
				if tC.name == "no events" {
					s.Require().Empty(result, "expected empty result, got non-empty")
				} else {
					castedEvents := make([]*types.Event, len(events))
					for i := range events {
						castedEvents[i] = events[i].ToEvent()
					}
					s.Require().Equal(castedEvents, result, "events do not match the expected ones")
				}
			}
		})
	}
}

func (s *SQLSuite) TestGetEventsForPeriod() {
	userID := "user1"
	start := time.Now()
	end := start.Add(24 * time.Hour)
	events := []*types.DBEvent{
		s.newTestEvent("Event 1", userID).ToDBEvent(),
		s.newTestEvent("Event 2", userID).ToDBEvent(),
	}
	userIDPtr := &userID

	testCases := []struct {
		name     string
		userID   *string
		dbMockFn func()
		txMockFn func()
		expected error
	}{
		{
			name:   "valid get with user",
			userID: userIDPtr,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockGetEvents(&events, true, 3)
				s.mockCommit(true)
			},
			expected: nil,
		},
		{
			name:   "valid get without user",
			userID: nil,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockGetEvents(&events, true, 3)
				s.mockCommit(true)
			},
			expected: nil,
		},
		{
			name:   "no events with user",
			userID: userIDPtr,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockGetEvents(&events, false, 3)
				s.mockCommit(true)
			},
			expected: projectErrors.ErrEventNotFound,
		},
		{
			name:   "no events without user",
			userID: nil,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.mockGetEvents(&events, false, 3)
				s.mockCommit(true)
			},
			expected: projectErrors.ErrEventNotFound,
		},
		{
			name:   "query error with user",
			userID: userIDPtr,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.txMock.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).
					Return(errUnknownErr).Once()
				s.mockRollback(true)
			},
			expected: projectErrors.ErrQeuryError,
		},
		{
			name:   "query error without user",
			userID: nil,
			dbMockFn: func() {
				s.mockBeginTx(true)
			},
			txMockFn: func() {
				s.txMock.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).
					Return(errUnknownErr).Once()
				s.mockRollback(true)
			},
			expected: projectErrors.ErrQeuryError,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.dbMockFn()
			tC.txMockFn()
			result, err := s.storage.GetEventsForPeriod(s.ctx, start, end, tC.userID)
			if tC.expected != nil {
				s.Require().Error(err, "expected error, got nil")
				s.Require().ErrorIs(err, tC.expected, "expected error does not match")
				s.Require().Nil(result, "expected nil result, got non-nil")
			} else {
				s.Require().NoError(err, "expected nil, got error")
				if tC.name == "no events with user" || tC.name == "no events without user" {
					s.Require().Empty(result, "expected empty result, got non-empty")
				} else {
					castedEvents := make([]*types.Event, len(events))
					for i := range events {
						castedEvents[i] = events[i].ToEvent()
					}
					s.Require().Equal(castedEvents, result, "events do not match the expected ones")
				}
			}
		})
	}
}
