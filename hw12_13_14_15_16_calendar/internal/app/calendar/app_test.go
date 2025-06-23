//nolint:depguard,nolintlint
package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/app/calendar/mocks" //nolint:depguard,nolintlint
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors"    //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"                   //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                             //nolint:depguard,nolintlint
	"github.com/stretchr/testify/mock"                                                   //nolint:depguard,nolintlint
	"github.com/stretchr/testify/require"                                                //nolint:depguard,nolintlint
)

func TestNewApp(t *testing.T) {
	testCases := []struct {
		name        string
		logger      Logger
		storage     Storage
		config      map[string]any
		expectedErr error
	}{
		{
			name:    "nil logger",
			logger:  nil,
			storage: &mocks.Storage{},
			config: map[string]any{
				"retries":       3,
				"retry_timeout": time.Second,
			},
			expectedErr: projectErrors.ErrAppInitFailed,
		},
		{
			name:    "nil storage",
			logger:  &mocks.Logger{},
			storage: nil,
			config: map[string]any{
				"retries":       3,
				"retry_timeout": time.Second,
			},
			expectedErr: projectErrors.ErrAppInitFailed,
		},
		{
			name:        "nil config",
			logger:      &mocks.Logger{},
			storage:     &mocks.Storage{},
			config:      nil,
			expectedErr: projectErrors.ErrAppInitFailed,
		},
		{
			name:    "missing retries in config",
			logger:  &mocks.Logger{},
			storage: &mocks.Storage{},
			config: map[string]any{
				"retry_timeout": time.Second,
			},
			expectedErr: projectErrors.ErrCorruptedConfig,
		},
		{
			name:    "wrong retries type in config",
			logger:  &mocks.Logger{},
			storage: &mocks.Storage{},
			config: map[string]any{
				"retries":       "not an int",
				"retry_timeout": time.Second,
			},
			expectedErr: projectErrors.ErrCorruptedConfig,
		},
		{
			name:    "zero retry timeout",
			logger:  &mocks.Logger{},
			storage: &mocks.Storage{},
			config: map[string]any{
				"retries":       3,
				"retry_timeout": time.Duration(0),
			},
			expectedErr: projectErrors.ErrCorruptedConfig,
		},
		{
			name:    "valid config",
			logger:  &mocks.Logger{},
			storage: &mocks.Storage{},
			config: map[string]any{
				"retries":       3,
				"retry_timeout": time.Millisecond * 100,
			},
			expectedErr: nil,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			app, err := NewApp(tC.logger, tC.storage, tC.config)

			if tC.expectedErr != nil {
				require.Error(t, err, "expected error, got nil")
				require.ErrorIs(t, err, tC.expectedErr, "unexpected error type")
				require.Nil(t, app, "app should be nil on error")
			} else {
				require.NoError(t, err, "expected nil, got error")
				require.NotNil(t, app, "app should not be nil")
			}
		})
	}
}

func TestGetEvents_RetryLogic(t *testing.T) {
	t.Run("success/retryable error", retryableSuccess)
	t.Run("success/retryable storage error", retryableStorageErrorSuccess)
	t.Run("failure/retry limit exceeded", retryableLimitExceeded)
	t.Run("failure/non retryable", nonRetryable)
}

func retryableSuccess(t *testing.T) {
	t.Helper()
	logger := new(mocks.Logger)
	storage := new(mocks.Storage)

	// Any retriable error is expected. Expecting 2 failed attempts.
	first := storage.On("GetEvent", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, projectErrors.ErrTimeoutExceeded).
		Twice()

	// Expecting 1 successful attempt.
	second := storage.On("GetEvent", mock.Anything, mock.Anything, mock.Anything).
		Return(&types.Event{EventData: types.EventData{Title: "Recovered"}}, nil).
		Once()

	// Expecting 2 debug logs - 1 for each failed attempt.
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return().Twice()

	second.NotBefore(first)

	app := &App{
		s:            storage,
		l:            logger,
		retries:      2, // 2 retries beside the first call.
		retryTimeout: time.Millisecond * 100,
	}

	event, err := app.GetEvent(context.Background(), uuid.New().String())

	require.NoError(t, err, "method should return no error after retries")
	require.NotNil(t, event, "event should not be nil")
	require.Equal(t, "Recovered", event.Title, "event title should match")

	storage.AssertNumberOfCalls(t, "GetEvent", 3)
}

func retryableStorageErrorSuccess(t *testing.T) {
	t.Helper()
	logger := new(mocks.Logger)
	storage := new(mocks.Storage)

	// Expecting the following sequence of calls on Storage.
	storageMocks := []*mock.Call{
		storage.On("GetEvent", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, projectErrors.ErrStorageUninitialized).
			Once(),
		storage.On("Connect", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).
			Once(),
		storage.On("GetEvent", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, projectErrors.ErrStorageUninitialized).
			Once(),
		storage.On("Connect", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).
			Once(),
		storage.On("GetEvent", mock.Anything, mock.Anything, mock.Anything).
			Return(&types.Event{EventData: types.EventData{Title: "Recovered"}}, nil).
			Once(),
	}

	// Expecting the following logger mocks.
	loggerMocks := []*mock.Call{
		logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return().
			Once(),
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return().
			Twice(),
		logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return().
			Once(),
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return().
			Twice(),
	}

	for i := 0; i < len(storageMocks)-2; i++ {
		storageMocks[i+1].NotBefore(storageMocks[i])
	}
	for i := 0; i < len(loggerMocks)-2; i++ {
		loggerMocks[i+1].NotBefore(loggerMocks[i])
	}

	app := &App{
		s:            storage,
		l:            logger,
		retries:      2, // 2 retries beside the first call.
		retryTimeout: time.Millisecond * 100,
	}

	event, err := app.GetEvent(context.Background(), uuid.New().String())

	require.NoError(t, err, "method should return no error after retries")
	require.NotNil(t, event, "event should not be nil")
	require.Equal(t, "Recovered", event.Title, "event title should match")

	storage.AssertNumberOfCalls(t, "GetEvent", 3)
}

func retryableLimitExceeded(t *testing.T) {
	t.Helper()
	logger := new(mocks.Logger)
	storage := new(mocks.Storage)

	// Any retriable error is expected. Expecting 3 failed attempts.
	storage.On("GetEvent", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, projectErrors.ErrTimeoutExceeded).
		Times(3)

	// Expecting 3 debug logs - 1 for each failed attempt. + error log.
	first := logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return().Times(3)
	second := logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return().Once()
	second.NotBefore(first)

	app := &App{
		s:            storage,
		l:            logger,
		retries:      2, // 2 retries beside the first call.
		retryTimeout: time.Millisecond * 100,
	}

	event, err := app.GetEvent(context.Background(), uuid.New().String())

	require.Error(t, err, "got nil, expected error")
	require.Nil(t, event, "not expecting event data")

	storage.AssertNumberOfCalls(t, "GetEvent", 3)
}

func nonRetryable(t *testing.T) {
	t.Helper()
	logger := new(mocks.Logger)
	storage := new(mocks.Storage)

	// Not retriable error is expected. Expecting 1 failed attempt.
	// This method really can't return such error, but nevermind.
	storage.On("GetEvent", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("some unexpected error")).
		Once()

	logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return().Once()

	app := &App{
		s:            storage,
		l:            logger,
		retries:      2, // 2 retries beside the first call.
		retryTimeout: time.Millisecond * 100,
	}

	event, err := app.GetEvent(context.Background(), uuid.New().String())

	require.Error(t, err, "got nil, expected error")
	require.Nil(t, event, "not expecting event data")

	storage.AssertNumberOfCalls(t, "GetEvent", 1)
}
