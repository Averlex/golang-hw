package memory_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors"              //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"               //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                         //nolint:depguard,nolintlint
	"github.com/stretchr/testify/require"                                            //nolint:depguard,nolintlint
	"github.com/stretchr/testify/suite"                                              //nolint:depguard,nolintlint
)

// canceledContext returns a canceled context for testing.
func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestNewStorage(t *testing.T) {
	testCases := []struct {
		name string
		size int
	}{
		{"positive size", 500},
		{"zero size", 0},
		{"negative size", -1},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			storage, err := memory.NewStorage(tC.size)
			require.NoError(t, err, "expected nil, got error")
			require.NotNil(t, storage, "expected non-nil storage, got nil")
		})
	}
}

func TestConnect(t *testing.T) {
	testCases := []struct {
		name      string
		ctx       context.Context
		withError bool
	}{
		{"successful connect", context.Background(), false},
		{"canceled context", canceledContext(), true},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			storage, err := memory.NewStorage(1000)
			require.NoError(t, err, "expected nil, got error")

			err = storage.Connect(tC.ctx)
			if tC.withError {
				require.Error(t, err, "expected error, got nil")
				require.ErrorIs(t, err, context.Canceled, "unexpected error type")
			} else {
				require.NoError(t, err, "expected nil, got error")
			}
		})
	}
}

func TestClose(t *testing.T) {
	testCases := []struct {
		name       string
		connect    bool
		closeTimes int
	}{
		{"close after connect", true, 1},
		{"close without connect", false, 1},
		{"consecutive close calls", true, 2},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			storage, err := memory.NewStorage(1000)
			require.NoError(t, err, "expected nil, got error")

			if tC.connect {
				err = storage.Connect(context.Background())
				require.NoError(t, err, "Connect should not return an error")
			}

			for i := 0; i < tC.closeTimes; i++ {
				storage.Close(context.Background())
			}
		})
	}
}

type StorageSuite struct {
	suite.Suite
	defaultStorageSize int
	eventTitle         string
	eventDuration      time.Duration
	eventDescription   string
	eventRemindIn      time.Duration
	userID, altUserID  string
}

func (s *StorageSuite) SetupSuite() {
	s.defaultStorageSize = 1000
	s.eventTitle = "Test Event"
	s.eventDuration = time.Hour
	s.eventDescription = "Description"
	s.eventRemindIn = time.Minute * 30
	s.userID = "user1"
	s.altUserID = "user2"
}

func TestStorage(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}

func (s *StorageSuite) createValidEvent() *types.Event {
	event, err := types.NewEvent(
		s.eventTitle,
		time.Now(),
		s.eventDuration,
		s.eventDescription,
		s.userID,
		s.eventRemindIn,
	)
	s.Require().NoError(err, "failed to create valid event")
	return event
}

func (s *StorageSuite) createValidEventData() *types.EventData {
	data, err := types.NewEventData(
		s.eventTitle,
		time.Now(),
		s.eventDuration,
		s.eventDescription,
		s.userID,
		s.eventRemindIn,
	)
	s.Require().NoError(err, "failed to create valid event")
	return data
}

//nolint:gocognit
func (s *StorageSuite) TestCreateEvent() {
	repeatedID := uuid.New()
	testCases := []struct {
		name          string
		event         *types.Event
		ctx           context.Context
		connect       bool
		prepare       func(*memory.Storage)
		withError     bool
		expectedError error
	}{
		{
			name:      "successful create",
			event:     s.createValidEvent(),
			ctx:       context.Background(),
			connect:   true,
			withError: false,
		},
		{
			name:          "nil event",
			event:         nil,
			ctx:           context.Background(),
			connect:       true,
			withError:     true,
			expectedError: errors.ErrNoData,
		},
		{
			name:          "uninitialized storage",
			event:         s.createValidEvent(),
			ctx:           context.Background(),
			connect:       false,
			withError:     true,
			expectedError: errors.ErrStorageUninitialized,
		},
		{
			name:    "event ID collision",
			event:   s.createValidEvent(),
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = repeatedID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for ID collision")
			},
			withError:     true,
			expectedError: errors.ErrDataExists,
		},
		{
			name:    "storage full",
			event:   s.createValidEvent(),
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				err := storage.Connect(context.Background())
				s.Require().NoError(err, "failed to connect storage")
				event := s.createValidEvent()
				_, err = storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to fill storage")
			},
			withError:     true,
			expectedError: errors.ErrStorageFull,
		},
		{
			name:    "date busy",
			event:   s.createValidEvent(),
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for date busy")
			},
			withError:     true,
			expectedError: errors.ErrDateBusy,
		},
		{
			name:          "canceled context",
			event:         s.createValidEvent(),
			ctx:           canceledContext(),
			connect:       true,
			withError:     true,
			expectedError: errors.ErrTimeoutExceeded,
		},
		{
			name:    "concurrent create",
			event:   s.createValidEvent(),
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				var wg sync.WaitGroup
				errCh := make(chan error, 10)
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						event := s.createValidEvent()
						event.Duration = time.Nanosecond
						time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
						_, err := storage.CreateEvent(context.Background(), event)
						if err != nil {
							errCh <- err
						}
					}()
				}
				wg.Wait()
				close(errCh)
				for err := range errCh {
					if err != nil {
						s.Require().ErrorIs(err, errors.ErrDateBusy, "unexpected error in concurrent create: %w", err)
					}
				}
			},
			withError: false,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			storage, err := memory.NewStorage(s.defaultStorageSize)
			if tC.name == "storage full" {
				storage, err = memory.NewStorage(1) // Create storage with size 1.
			}
			s.Require().NoError(err, "expected nil, got error on NewStorage")

			if tC.name == "event ID collision" {
				tC.event.ID = repeatedID // Set the ID to a repeated one.
			}
			if tC.name == "concurrent create" {
				tC.event.Duration = time.Nanosecond // Set the duration to a short one.
			}

			if tC.connect {
				err = storage.Connect(context.Background())
				s.Require().NoError(err, "expected nil, got error on Connect")
			}

			if tC.prepare != nil {
				tC.prepare(storage)
			}

			result, err := storage.CreateEvent(tC.ctx, tC.event)
			if tC.withError {
				s.Require().Error(err, "expected error, got nil")
				if tC.expectedError != nil {
					s.Require().ErrorIs(err, tC.expectedError, "unexpected error type")
				}
			} else {
				s.Require().NoError(err, "expected nil, got error")
				s.Require().NotNil(result, "expected non-nil event, got nil")
				s.Require().Equal(tC.event, result, "returned event does not match input")
			}
		})
	}
}

//nolint:gocognit,funlen
func (s *StorageSuite) TestUpdateEvent() {
	existingEventID := uuid.New()
	testCases := []struct {
		name          string
		eventID       uuid.UUID
		data          *types.EventData
		ctx           context.Context
		connect       bool
		prepare       func(*memory.Storage)
		withError     bool
		expectedError error
	}{
		{
			name:    "successful update",
			eventID: existingEventID,
			data:    s.createValidEventData(),
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for update")
			},
			withError: false,
		},
		{
			name:          "nil data",
			eventID:       existingEventID,
			data:          nil,
			ctx:           context.Background(),
			connect:       true,
			withError:     true,
			expectedError: errors.ErrNoData,
		},
		{
			name:          "uninitialized storage",
			eventID:       existingEventID,
			data:          s.createValidEventData(),
			ctx:           context.Background(),
			connect:       false,
			withError:     true,
			expectedError: errors.ErrStorageUninitialized,
		},
		{
			name:          "event not found",
			eventID:       uuid.New(),
			data:          s.createValidEventData(),
			ctx:           context.Background(),
			connect:       true,
			withError:     true,
			expectedError: errors.ErrEventNotFound,
		},
		{
			name:    "permission denied",
			eventID: existingEventID,
			data: func() *types.EventData {
				data := s.createValidEventData()
				data.UserID = s.altUserID
				return data
			}(),
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for permission denied")
			},
			withError:     true,
			expectedError: errors.ErrPermissionDenied,
		},
		{
			name:    "date busy",
			eventID: existingEventID,
			data:    s.createValidEventData(),
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event1 := s.createValidEvent()
				event2 := s.createValidEvent()
				event2.ID = existingEventID                      // Atm events are not overlapping each other.
				event2.Datetime = time.Now().Add(24 * time.Hour) // Will move event2 on event1 in the test.
				_, err := storage.CreateEvent(context.Background(), event1)
				s.Require().NoError(err, "failed to prepare event1 for date busy")
				_, err = storage.CreateEvent(context.Background(), event2)
				s.Require().NoError(err, "failed to prepare event2 for date busy")
			},
			withError:     true,
			expectedError: errors.ErrDateBusy,
		},
		{
			name:    "canceled context",
			eventID: existingEventID,
			data:    s.createValidEventData(),
			ctx:     canceledContext(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for canceled context")
			},
			withError:     true,
			expectedError: errors.ErrTimeoutExceeded,
		},
		{
			name:    "concurrent update",
			eventID: existingEventID,
			data:    s.createValidEventData(),
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for concurrent update")
				var wg sync.WaitGroup
				errCh := make(chan error, 10)
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						data := s.createValidEventData()
						data.Datetime = time.Now().Add(time.Duration(i+1) * time.Hour)
						_, err := storage.UpdateEvent(context.Background(), existingEventID, data)
						if err != nil {
							errCh <- err
						}
					}()
				}
				wg.Wait()
				close(errCh)
				for err := range errCh {
					if err != nil {
						s.Require().ErrorIs(err, errors.ErrDateBusy, "unexpected error in concurrent update: %w", err)
					}
				}
			},
			withError: false,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			storage, err := memory.NewStorage(s.defaultStorageSize)
			s.Require().NoError(err, "expected nil, got error on NewStorage")

			if tC.connect {
				err = storage.Connect(context.Background())
				s.Require().NoError(err, "expected nil, got error on Connect")
			}

			if tC.prepare != nil {
				tC.prepare(storage)
			}

			result, err := storage.UpdateEvent(tC.ctx, tC.eventID, tC.data)
			if tC.withError {
				s.Require().Error(err, "expected error, got nil")
				if tC.expectedError != nil {
					s.Require().ErrorIs(err, tC.expectedError, "unexpected error type")
				}
			} else {
				s.Require().NoError(err, "expected nil, got error")
				s.Require().NotNil(result, "expected non-nil event, got nil")
				s.Require().Equal(tC.eventID, result.ID, "event ID mismatch")
				s.Require().Equal(tC.data.UserID, result.UserID, "user ID mismatch")
				s.Require().Equal(tC.data.Title, result.Title, "title mismatch")
				s.Require().Equal(tC.data.Datetime, result.Datetime, "datetime mismatch")
				s.Require().Equal(tC.data.Duration, result.Duration, "duration mismatch")
				if tC.data.Description != nil {
					s.Require().Equal(*tC.data.Description, *result.Description, "description mismatch")
				}
				if tC.data.RemindIn != nil {
					s.Require().Equal(*tC.data.RemindIn, *result.RemindIn, "remindIn mismatch")
				}
			}
		})
	}
}

func (s *StorageSuite) TestDeleteEvent() {
	existingEventID := uuid.New()
	testCases := []struct {
		name          string
		eventID       uuid.UUID
		ctx           context.Context
		connect       bool
		prepare       func(*memory.Storage)
		withError     bool
		expectedError error
	}{
		{
			name:    "successful delete",
			eventID: existingEventID,
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for delete")
			},
			withError: false,
		},
		{
			name:          "uninitialized storage",
			eventID:       existingEventID,
			ctx:           context.Background(),
			connect:       false,
			withError:     true,
			expectedError: errors.ErrStorageUninitialized,
		},
		{
			name:          "event not found",
			eventID:       uuid.New(),
			ctx:           context.Background(),
			connect:       true,
			withError:     true,
			expectedError: errors.ErrEventNotFound,
		},
		{
			name:    "canceled context",
			eventID: existingEventID,
			ctx:     canceledContext(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for canceled context")
			},
			withError:     true,
			expectedError: errors.ErrTimeoutExceeded,
		},
		{
			name:    "concurrent delete",
			eventID: existingEventID,
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				events := make([]*types.Event, 0, 5)
				// Creating 10 valid events to ensure the storage is not empty.
				for i := 0; i < 10; i++ {
					events = append(events, s.createValidEvent())
					events[i].Duration = time.Nanosecond
					events[i].Datetime = time.Now().Add(time.Duration(i) * time.Hour)
				}
				var wg sync.WaitGroup
				errCh := make(chan error, 10)
				// Attempting to delete each element concurrently.
				for i := 0; i < 10; i++ {
					_, err := storage.CreateEvent(context.Background(), events[i])
					s.Require().NoError(err, "failed to prepare event for concurrent delete: %w", err)
					wg.Add(1)
					go func(elem *types.Event) {
						defer wg.Done()
						err := storage.DeleteEvent(context.Background(), elem.ID)
						if err != nil {
							errCh <- err
						}
					}(events[i])
				}
				wg.Wait()
				close(errCh)
				for err := range errCh {
					if err != nil {
						s.Require().NoError(err, "unexpected error in concurrent delete: %w", err)
					}
				}
				// For upwards compatibility, we ensure that the event with existingEventID is still present.
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for concurrent delete")
			},
			withError: false,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			storage, err := memory.NewStorage(s.defaultStorageSize)
			s.Require().NoError(err, "expected nil, got error on NewStorage")

			if tC.connect {
				err = storage.Connect(context.Background())
				s.Require().NoError(err, "expected nil, got error on Connect")
			}

			if tC.prepare != nil {
				tC.prepare(storage)
			}

			err = storage.DeleteEvent(tC.ctx, tC.eventID)
			if tC.withError {
				s.Require().Error(err, "expected error, got nil")
				if tC.expectedError != nil {
					s.Require().ErrorIs(err, tC.expectedError, "unexpected error type")
				}
			} else {
				s.Require().NoError(err, "expected nil, got error")
			}
		})
	}
}

func (s *StorageSuite) TestSequentialOperations() {
	s.Run("sequential create-update-delete", func() {
		storage, err := memory.NewStorage(s.defaultStorageSize)
		s.Require().NoError(err, "expected nil, got error on NewStorage")

		err = storage.Connect(context.Background())
		s.Require().NoError(err, "expected nil, got error on Connect")

		// Create event
		event := s.createValidEvent()
		result, err := storage.CreateEvent(context.Background(), event)
		s.Require().NoError(err, "expected nil, got error on CreateEvent")
		s.Require().NotNil(result, "expected non-nil event, got nil")
		s.Require().Equal(event, result, "created event does not match input")

		// Update event
		data := s.createValidEventData()
		data.Datetime = time.Now().Add(2 * time.Hour) // Ensure no overlap
		updated, err := storage.UpdateEvent(context.Background(), event.ID, data)
		s.Require().NoError(err, "expected nil, got error on UpdateEvent")
		s.Require().NotNil(updated, "expected non-nil event, got nil")
		s.Require().Equal(event.ID, updated.ID, "event ID mismatch")
		s.Require().Equal(data.UserID, updated.UserID, "user ID mismatch")
		s.Require().Equal(data.Title, updated.Title, "title mismatch")
		s.Require().Equal(data.Datetime, updated.Datetime, "datetime mismatch")

		// Delete event
		err = storage.DeleteEvent(context.Background(), event.ID)
		s.Require().NoError(err, "expected nil, got error on DeleteEvent")

		// Verify event is deleted
		_, err = storage.UpdateEvent(context.Background(), event.ID, data)
		s.Require().Error(err, "expected error, got nil on UpdateEvent after delete")
		s.Require().ErrorIs(err, errors.ErrEventNotFound, "unexpected error type")
	})
}

func (s *StorageSuite) TestStorageSizeLimits() {
	testCases := []struct {
		name          string
		storageSize   int
		eventCount    int
		withError     bool
		expectedError error
		withDelete    bool
	}{
		{
			name:        "small storage success",
			storageSize: 2,
			eventCount:  2,
			withError:   false,
		},
		{
			name:          "storage full",
			storageSize:   1,
			eventCount:    2,
			withError:     true,
			expectedError: errors.ErrStorageFull,
		},
		{
			name:        "zero storage size",
			storageSize: 0,
			eventCount:  1,
			withError:   false, // Should use default size (1000).
		},
		{
			name:        "stress test",
			storageSize: 100_000,
			eventCount:  100_000,
			withError:   false,
			withDelete:  true,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			storage, err := memory.NewStorage(tC.storageSize)
			s.Require().NoError(err, "expected nil, got error on NewStorage")

			err = storage.Connect(context.Background())
			s.Require().NoError(err, "expected nil, got error on Connect")

			for i := 0; i < tC.eventCount; i++ {
				event := s.createValidEvent()
				event.Datetime = time.Now().Add(time.Duration(i+1) * time.Hour) // Avoid overlaps
				_, err = storage.CreateEvent(context.Background(), event)
				if i < tC.eventCount-1 || !tC.withError {
					s.Require().NoError(err, "expected nil, got error on CreateEvent: %w", err)
				} else {
					s.Require().Error(err, "expected error, got nil on CreateEvent")
					s.Require().ErrorIs(err, tC.expectedError, "unexpected error type")
				}

				if tC.withDelete {
					err = storage.DeleteEvent(context.Background(), event.ID)
					s.Require().NoError(err, "expected nil, got error on DeleteEvent")
				}
			}
		})
	}
}

func (s *StorageSuite) TestGetEvent() {
	existingEventID := uuid.New()
	testCases := []struct {
		name          string
		eventID       uuid.UUID
		ctx           context.Context
		connect       bool
		prepare       func(*memory.Storage)
		withError     bool
		expectedError error
	}{
		{
			name:    "successful get",
			eventID: existingEventID,
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for get")
			},
			withError: false,
		},
		{
			name:          "event not found",
			eventID:       uuid.New(),
			ctx:           context.Background(),
			connect:       true,
			withError:     true,
			expectedError: errors.ErrEventNotFound,
		},
		{
			name:          "uninitialized storage",
			eventID:       existingEventID,
			ctx:           context.Background(),
			connect:       false,
			withError:     true,
			expectedError: errors.ErrStorageUninitialized,
		},
		{
			name:    "canceled context",
			eventID: existingEventID,
			ctx:     canceledContext(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for canceled context")
			},
			withError:     true,
			expectedError: errors.ErrTimeoutExceeded,
		},
		{
			name:    "concurrent get",
			eventID: existingEventID,
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for concurrent get")
				var wg sync.WaitGroup
				errCh := make(chan error, 10)
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						_, err := storage.GetEvent(context.Background(), existingEventID)
						if err != nil {
							errCh <- err
						}
					}()
				}
				wg.Wait()
				close(errCh)
				for err := range errCh {
					if err != nil {
						s.Require().NoError(err, "unexpected error in concurrent get: %w", err)
					}
				}
			},
			withError: false,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			storage, err := memory.NewStorage(s.defaultStorageSize)
			s.Require().NoError(err, "expected nil, got error on NewStorage")

			if tC.connect {
				err = storage.Connect(context.Background())
				s.Require().NoError(err, "expected nil, got error on Connect")
			}

			if tC.prepare != nil {
				tC.prepare(storage)
			}

			result, err := storage.GetEvent(tC.ctx, tC.eventID)
			if tC.withError {
				s.Require().Error(err, "expected error, got nil")
				s.Require().Nil(result, "expected nil event, got non-nil")
				if tC.expectedError != nil {
					s.Require().ErrorIs(err, tC.expectedError, "unexpected error type")
				}
			} else {
				s.Require().NoError(err, "expected nil, got error")
				s.Require().NotNil(result, "expected non-nil event, got nil")
				s.Require().Equal(tC.eventID, result.ID, "event ID mismatch")
			}
		})
	}
}

func (s *StorageSuite) TestGetAllUserEvents() {
	userID := s.userID
	nonExistentUserID := s.altUserID
	existingEventID := uuid.New()
	testCases := []struct {
		name          string
		userID        string
		ctx           context.Context
		connect       bool
		prepare       func(*memory.Storage)
		withError     bool
		expectedError error
		eventCount    int
	}{
		{
			name:    "successful get",
			userID:  userID,
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for get")
			},
			withError:  false,
			eventCount: 1,
		},
		{
			name:          "no events for user",
			userID:        nonExistentUserID,
			ctx:           context.Background(),
			connect:       true,
			withError:     true,
			expectedError: errors.ErrEventNotFound,
		},
		{
			name:          "uninitialized storage",
			userID:        userID,
			ctx:           context.Background(),
			connect:       false,
			withError:     true,
			expectedError: errors.ErrStorageUninitialized,
		},
		{
			name:    "canceled context",
			userID:  userID,
			ctx:     canceledContext(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for canceled context")
			},
			withError:     true,
			expectedError: errors.ErrTimeoutExceeded,
		},
		{
			name:    "concurrent get",
			userID:  userID,
			ctx:     context.Background(),
			connect: true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.ID = existingEventID
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event for concurrent get")
				var wg sync.WaitGroup
				errCh := make(chan error, 10)
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						_, err := storage.GetAllUserEvents(context.Background(), userID)
						if err != nil {
							errCh <- err
						}
					}()
				}
				wg.Wait()
				close(errCh)
				for err := range errCh {
					if err != nil {
						s.Require().NoError(err, "unexpected error in concurrent get: %w", err)
					}
				}
			},
			withError:  false,
			eventCount: 1,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			storage, err := memory.NewStorage(s.defaultStorageSize)
			s.Require().NoError(err, "expected nil, got error on NewStorage")

			if tC.connect {
				err = storage.Connect(context.Background())
				s.Require().NoError(err, "expected nil, got error on Connect")
			}

			if tC.prepare != nil {
				tC.prepare(storage)
			}

			result, err := storage.GetAllUserEvents(tC.ctx, tC.userID)
			if tC.withError {
				s.Require().Error(err, "expected error, got nil")
				s.Require().Nil(result, "expected nil events, got non-nil")
				if tC.expectedError != nil {
					s.Require().ErrorIs(err, tC.expectedError, "unexpected error type")
				}
			} else {
				s.Require().NoError(err, "expected nil, got error")
				s.Require().NotNil(result, "expected non-nil events, got nil")
				s.Require().Len(result, tC.eventCount, "unexpected number of events")
				if tC.eventCount > 0 {
					s.Require().Equal(tC.userID, result[0].UserID, "user ID mismatch")
				}
			}
		})
	}
}

func (s *StorageSuite) TestGetEventsForPeriod() {
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := startDate.AddDate(0, 0, 2)
	userID := s.userID
	nonExistentUserID := s.altUserID
	var nilUserID *string

	testCases := []struct {
		name       string
		startDate  time.Time
		endDate    time.Time
		userID     *string
		ctx        context.Context
		connect    bool
		prepare    func(*memory.Storage)
		wantErr    error
		eventCount int
	}{
		{
			name:      "success for user",
			startDate: startDate,
			endDate:   endDate,
			userID:    &userID,
			ctx:       context.Background(),
			connect:   true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.Datetime = startDate.Add(time.Hour)
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event")
			},
			eventCount: 1,
		},
		{
			name:      "success for all users",
			startDate: startDate,
			endDate:   endDate,
			userID:    nilUserID,
			ctx:       context.Background(),
			connect:   true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.Datetime = startDate.Add(time.Hour)
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event")
			},
			eventCount: 1,
		},
		{
			name:      "no events in period",
			startDate: startDate.AddDate(0, 0, 10),
			endDate:   endDate.AddDate(0, 0, 10),
			userID:    &userID,
			ctx:       context.Background(),
			connect:   true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event")
			},
			wantErr: errors.ErrEventNotFound,
		},
		{
			name:      "no events for user",
			startDate: startDate,
			endDate:   endDate,
			userID:    &nonExistentUserID,
			ctx:       context.Background(),
			connect:   true,
			wantErr:   errors.ErrEventNotFound,
		},
		{
			name:      "uninitialized storage",
			startDate: startDate,
			endDate:   endDate,
			userID:    &userID,
			ctx:       context.Background(),
			connect:   false,
			wantErr:   errors.ErrStorageUninitialized,
		},
		{
			name:      "canceled context",
			startDate: startDate,
			endDate:   endDate,
			userID:    &userID,
			ctx:       canceledContext(),
			connect:   true,
			prepare: func(storage *memory.Storage) {
				event := s.createValidEvent()
				event.Datetime = startDate.Add(time.Hour)
				_, err := storage.CreateEvent(context.Background(), event)
				s.Require().NoError(err, "failed to prepare event")
			},
			wantErr: errors.ErrTimeoutExceeded,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			storage, err := memory.NewStorage(s.defaultStorageSize)
			s.Require().NoError(err, "failed to create storage")

			if tC.connect {
				err = storage.Connect(context.Background())
				s.Require().NoError(err, "failed to connect storage")
			}

			if tC.prepare != nil {
				tC.prepare(storage)
			}

			result, err := storage.GetEventsForPeriod(tC.ctx, tC.startDate, tC.endDate, tC.userID)
			if tC.wantErr != nil {
				s.Require().ErrorIs(err, tC.wantErr, "unexpected error")
				s.Require().Nil(result, "expected nil events")
			} else {
				s.Require().NoError(err, "unexpected error")
				s.Require().Len(result, tC.eventCount, "wrong event count")
				if tC.userID != nil && len(result) > 0 {
					s.Require().Equal(*tC.userID, result[0].UserID, "user ID mismatch")
				}
			}
		})
	}
}

func (s *StorageSuite) TestGetEventsForPeriod_VariousSizes() {
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := startDate.AddDate(0, 0, 10)
	userID := s.userID
	var nilUserID *string

	// Prepare storage with multiple events
	storage, err := memory.NewStorage(s.defaultStorageSize)
	s.Require().NoError(err, "failed to create storage")
	err = storage.Connect(context.Background())
	s.Require().NoError(err, "failed to connect storage")

	// Add 5 events for user
	eventDates := []time.Time{
		startDate.Add(time.Hour),
		startDate.Add(2 * time.Hour),
		startDate.Add(3 * time.Hour),
		startDate.Add(4 * time.Hour),
		startDate.Add(5 * time.Hour),
	}
	for _, date := range eventDates {
		event := s.createValidEvent()
		event.Datetime = date
		_, err := storage.CreateEvent(context.Background(), event)
		s.Require().NoError(err, "failed to create event")
	}

	// Add 1 event for another user
	anotherUserID := s.altUserID
	event := s.createValidEvent()
	event.UserID = anotherUserID
	event.Datetime = startDate.Add(time.Hour)
	_, err = storage.CreateEvent(context.Background(), event)
	s.Require().NoError(err, "failed to create event")

	testCases := []struct {
		name       string
		startDate  time.Time
		endDate    time.Time
		userID     *string
		eventCount int
	}{
		{
			name:       "single event for user",
			startDate:  startDate,
			endDate:    startDate.Add(90 * time.Minute),
			userID:     &userID,
			eventCount: 1,
		},
		{
			name:       "three events for user",
			startDate:  startDate,
			endDate:    startDate.Add(4 * time.Hour),
			userID:     &userID,
			eventCount: 3,
		},
		{
			name:       "all user events",
			startDate:  startDate.Add(-time.Hour),
			endDate:    startDate.Add(6 * time.Hour),
			userID:     &userID,
			eventCount: 5,
		},
		{
			name:       "all storage events",
			startDate:  startDate.Add(-time.Hour),
			endDate:    startDate.Add(6 * time.Hour),
			userID:     nilUserID,
			eventCount: 6,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			result, err := storage.GetEventsForPeriod(context.Background(), tC.startDate, tC.endDate, tC.userID)
			s.Require().NoError(err, "unexpected error")
			s.Require().Len(result, tC.eventCount, "wrong event count")
			for _, event := range result {
				if tC.userID != nil {
					s.Require().Equal(*tC.userID, event.UserID, "user ID mismatch")
				}
				s.Require().False(event.Datetime.Before(tC.startDate) || event.Datetime.After(tC.endDate),
					"event datetime outside period")
			}
		})
	}

	// Concurrent access test
	s.Run("concurrent get", func() {
		var wg sync.WaitGroup
		errCh := make(chan error, 10)
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := storage.GetEventsForPeriod(context.Background(), startDate, endDate, &userID)
				if err != nil {
					errCh <- err
				}
			}()
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			s.Require().NoError(err, "unexpected error in concurrent get: %s", err)
		}
	})
}
