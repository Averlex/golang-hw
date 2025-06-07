package storage_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage"         //nolint:depguard,nolintlint
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/stretchr/testify/require"                                             //nolint:depguard,nolintlint
)

func copyMap(m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		if subMap, ok := v.(map[string]any); ok {
			v = copyMap(subMap)
		}
		result[k] = v
	}
	return result
}

func TestNewStorage(t *testing.T) {
	defaultMemArgs := map[string]any{
		"type": "memory",
		"memory": map[string]any{
			"size": 100,
		},
	}

	defaultSQLArgs := map[string]any{
		"type": "sql",
		"sql": map[string]any{
			"driver":   "postgres",
			"host":     "localhost",
			"port":     "5432",
			"user":     "testuser",
			"password": "pass",
			"dbname":   "testdb",
			"timeout":  5 * time.Second,
		},
	}

	testCases := []struct {
		name          string
		args          map[string]any
		expectedError error
	}{
		{
			name:          "no config passed",
			args:          nil,
			expectedError: projectErrors.ErrCorruptedConfig,
		},
		{
			name:          "empty config",
			args:          map[string]any{},
			expectedError: projectErrors.ErrCorruptedConfig,
		},
		{
			name:          "missing type",
			args:          map[string]any{"memory": map[string]any{}},
			expectedError: projectErrors.ErrCorruptedConfig,
		},
		{
			name:          "unknown storage type",
			args:          map[string]any{"type": "invalid"},
			expectedError: projectErrors.ErrCorruptedConfig,
		},
		{
			name:          "memory/valid",
			args:          defaultMemArgs,
			expectedError: nil,
		},
		{
			name:          "sql/valid",
			args:          defaultSQLArgs,
			expectedError: nil,
		},
		{
			name: "sql/alternative driver",
			args: func() map[string]any {
				cfg := copyMap(defaultSQLArgs)
				cfg["sql"].(map[string]any)["driver"] = "postgresql"
				return cfg
			}(),
			expectedError: nil,
		},
		{
			name: "memory/missing size",
			args: func() map[string]any {
				cfg := copyMap(defaultMemArgs)
				delete(cfg["memory"].(map[string]any), "size")
				return cfg
			}(),
			expectedError: projectErrors.ErrCorruptedConfig,
		},
		{
			name: "memory/invalid size type",
			args: func() map[string]any {
				cfg := copyMap(defaultMemArgs)
				cfg["memory"].(map[string]any)["size"] = "not_an_int"
				return cfg
			}(),
			expectedError: projectErrors.ErrCorruptedConfig,
		},
		{
			name: "sql/missing host",
			args: func() map[string]any {
				cfg := copyMap(defaultSQLArgs)
				delete(cfg["sql"].(map[string]any), "host")
				return cfg
			}(),
			expectedError: projectErrors.ErrCorruptedConfig,
		},
		{
			name: "sql/invalid timeout type",
			args: func() map[string]any {
				cfg := copyMap(defaultSQLArgs)
				cfg["sql"].(map[string]any)["timeout"] = "not_a_duration"
				return cfg
			}(),
			expectedError: projectErrors.ErrCorruptedConfig,
		},
		{
			name: "sql/driver not a string",
			args: func() map[string]any {
				cfg := copyMap(defaultSQLArgs)
				cfg["sql"].(map[string]any)["driver"] = 123
				return cfg
			}(),
			expectedError: projectErrors.ErrCorruptedConfig,
		},
		{
			name: "sql/usupported driver",
			args: func() map[string]any {
				cfg := copyMap(defaultSQLArgs)
				cfg["sql"].(map[string]any)["driver"] = "fictional driver"
				return cfg
			}(),
			expectedError: projectErrors.ErrStorageInitFailed,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			if tC.name == "sql/driver not a string" {
				fmt.Println()
			}

			s, err := storage.NewStorage(tC.args)

			if tC.expectedError == nil {
				require.NoError(t, err, "expected no error while creating storage")
				require.NotNil(t, s, "expected non-nil storage instance")
			} else {
				require.Error(t, err, "expected error while creating storage")
				require.ErrorIs(t, err, tC.expectedError, "unexpected error type")
			}
		})
	}
}
