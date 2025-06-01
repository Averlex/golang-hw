package memory_test

import (
	"context"
	"testing"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory" //nolint:depguard,nolintlint
	"github.com/stretchr/testify/require"                                            //nolint:depguard,nolintlint
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
		name        string
		ctx         context.Context
		expectError bool
	}{
		{"successful connect", context.Background(), false},
		{"canceled context", canceledContext(), true},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			storage, err := memory.NewStorage(1000)
			require.NoError(t, err, "expected nil, got error")

			err = storage.Connect(tC.ctx)
			if tC.expectError {
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
