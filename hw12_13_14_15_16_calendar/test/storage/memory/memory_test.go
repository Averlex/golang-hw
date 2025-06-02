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
	userID             string
}

func (s *StorageSuite) SetupSuite() {
	s.defaultStorageSize = 1000
	s.eventTitle = "Test Event"
	s.eventDuration = time.Hour
	s.eventDescription = "Description"
	s.eventRemindIn = time.Minute * 30
	s.userID = "user1"
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

//nolint:unused
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
