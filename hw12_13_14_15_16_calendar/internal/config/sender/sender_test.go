package sender

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require" //nolint:depguard,nolintlint
)

func TestGetSubConfig_Success(t *testing.T) {
	testCases := []struct {
		Name     string
		Config   Config
		Key      string
		Expected map[string]any
	}{
		{
			Name: "simple section",
			Config: Config{
				Logger: LoggerConf{
					Level:        "info",
					Format:       "json",
					TimeTemplate: time.UnixDate,
					LogStream:    "stdout",
				},
			},
			Key: "logger",
			Expected: map[string]any{
				"level":         "info",
				"format":        "json",
				"time_template": time.UnixDate,
				"log_stream":    "stdout",
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.Name, func(t *testing.T) {
			cfg := &tC.Config
			subCfg, err := cfg.GetSubConfig(tC.Key)
			require.NoError(t, err, "expected no error when extracting subconfig")

			require.True(t, reflect.DeepEqual(tC.Expected, subCfg),
				"expected and actual configs do not match")
		})
	}
}

func TestGetSubConfig_Error(t *testing.T) {
	testCases := []struct {
		Name   string
		Config Config
		Key    string
	}{
		{
			Name:   "key does not exist",
			Config: Config{},
			Key:    "nonexistent",
		},
		{
			Name: "key refers to non-struct field",
			Config: Config{
				Logger: LoggerConf{
					Level:        "info",
					Format:       "json",
					TimeTemplate: time.UnixDate,
					LogStream:    "stdout",
				},
			},
			Key: "log_stream",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.Name, func(t *testing.T) {
			cfg := &tC.Config
			_, err := cfg.GetSubConfig(tC.Key)
			require.Error(t, err)
		})
	}
}
