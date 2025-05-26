package logger

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type logEntry struct {
	Msg   string `json:"msg"`
	Level string `json:"level"`
	Time  string `json:"time"`
}

// customWriter is a log entries collector.
type customWriter struct {
	arr [][]byte
}

func (w *customWriter) Write(data []byte) (int, error) {
	copied := make([]byte, len(data))
	copy(copied, data)
	w.arr = append(w.arr, copied)
	return len(data), nil
}

func (w *customWriter) CleanUp() {
	w.arr = make([][]byte, 0)
}

func newCustomWriter() *customWriter {
	w := customWriter{}
	w.arr = make([][]byte, 0)
	return &w
}

func decodeJSON(data []byte) (*logEntry, error) {
	var buffer logEntry
	err := json.Unmarshal(data, &buffer)
	return &buffer, err
}

func TestLogger(t *testing.T) {
	t.Run("log level", logLevelTest)
	t.Run("log type", logTypeTest)
	t.Run("time template", timeTemplateTest)
}

func TestLogger_default(t *testing.T) {
	w := newCustomWriter()

	t.Run("nil writer", func(t *testing.T) {
		_, err := NewLogger("", "", "", nil)
		require.ErrorIs(t, err, ErrInvalidWriter, "unexpected error received: "+err.Error())
	})

	t.Run("empty log type", func(t *testing.T) {
		l, err := NewLogger("", "info", time.UnixDate, w) // JSON type expected.
		require.NoError(t, err, "got error, expected nil")

		l.Info("test")
		require.Len(t, w.arr, 1)

		entry, err := decodeJSON(w.arr[0])
		require.NoError(t, err, "got error, expected nil")

		require.Equal(t, "test", entry.Msg)
	})
	w.CleanUp()

	t.Run("empty log level", func(t *testing.T) {
		l, err := NewLogger("json", "", time.UnixDate, w) // ERROR level expected.
		require.NoError(t, err, "got error, expected nil")

		l.Debug("debug")
		l.Info("info")
		l.Warn("warn")
		l.Error("error")
		require.Len(t, w.arr, 1)

		entry, err := decodeJSON(w.arr[0])
		require.NoError(t, err, "got error, expected nil")

		require.Equal(t, "error", strings.ToLower(entry.Level))
	})
	w.CleanUp()

	t.Run("empty time template", func(t *testing.T) {
		l, err := NewLogger("json", "info", "", w) // "02.01.2006 15:04:05.000" time template expected.
		require.NoError(t, err, "got error, expected nil")

		l.Info("test")
		require.Len(t, w.arr, 1)

		entry, err := decodeJSON(w.arr[0])
		require.NoError(t, err, "got error, expected nil")

		logTime, err := time.ParseInLocation("02.01.2006 15:04:05.000", entry.Time, time.Local)
		require.NoError(t, err, "got error, expected nil")
		require.InDelta(t, time.Now().UnixMilli(), logTime.UnixMilli(), float64(500),
			"measured time doesn't match the logged one") // Comparing with an error margin of 500ms.
	})
	w.CleanUp()
}

func logLevelTest(t *testing.T) {
	t.Helper()
	w := newCustomWriter()

	callOrder := []string{"debug", "info", "warn", "error"}
	testCases := []struct {
		name         string
		level        string
		msg          string
		expectedSize int
	}{
		{"debug", "debug", "debug-test", 4},
		{"info", "info", "info-test", 3},
		{"warn", "warn", "warn-test", 2},
		{"error", "error", "error-test", 1},
		{"case insensitivity", "dEbUg", "dEbUg-test", 4},
		{"unknown", "unknown", "unknown-test", 0},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			w.CleanUp()
			l, err := NewLogger("json", tC.level, time.UnixDate, w)
			if tC.name == "unknown" {
				require.ErrorIs(t, err, ErrInvalidLogLevel, "unexpected error received: "+err.Error())
				return
			}
			require.NoError(t, err, "got error, expected nil")

			l.Debug(tC.msg)
			l.Info(tC.msg)
			l.Warn(tC.msg)
			l.Error(tC.msg)

			require.Len(t, w.arr, tC.expectedSize)
			for i := 0; i < len(w.arr); i++ {
				entry, err := decodeJSON(w.arr[i])
				require.NoError(t, err, "got error, expected nil")

				require.Equal(t, tC.msg, entry.Msg, "unexpected log message")
				require.Equal(t, callOrder[len(callOrder)-tC.expectedSize:][i], strings.ToLower(entry.Level),
					"unexpected log level")
			}
		})
	}
}

func logTypeTest(t *testing.T) {
	t.Helper()
	w := newCustomWriter()

	testCases := []struct {
		name                   string
		logType                string
		msg                    string
		expectConstructorError bool
		expectDecodingError    bool
	}{
		{"text", "text", "text-test", false, true},
		{"json", "json", "json-test", false, false},
		{"unknown", "unknown", "unknown-test", true, false},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			w.CleanUp()
			l, err := NewLogger(tC.logType, "info", time.UnixDate, w)
			if tC.expectConstructorError {
				require.ErrorIs(t, err, ErrInvalidLogType, "got nil, expected error")
				return
			}
			require.NoError(t, err, "got error, expected nil")

			l.Info(tC.msg)

			require.Len(t, w.arr, 1)
			_, err = decodeJSON(w.arr[0])
			if tC.expectDecodingError {
				require.Error(t, err, "got nil, expected error")
				return
			}
			require.NoError(t, err, "got error, expected nil")
		})
	}
}

func timeTemplateTest(t *testing.T) {
	t.Helper()
	w := newCustomWriter()

	testCases := []struct {
		name        string
		template    string
		expectedFmt string
		expectError bool
	}{
		{
			name:        "unix format",
			template:    time.UnixDate,
			expectedFmt: "Mon Jan _2 15:04:05 MST 2006",
			expectError: false,
		},
		{
			name:        "custom format",
			template:    "02.01.2006 15:04:05.000",
			expectedFmt: "02.01.2006 15:04:05.000",
			expectError: false,
		},
		{
			name:        "empty format",
			template:    "",
			expectedFmt: "02.01.2006 15:04:05.000",
			expectError: false,
		},
		{
			name:        "invalid format",
			template:    "invalid",
			expectError: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			w.CleanUp()
			if tC.name == "invalid format" {
				fmt.Println()
			}
			l, err := NewLogger("json", "info", tC.template, w)

			if tC.expectError {
				require.Error(t, err, "got nil, expected error")
				return
			}
			require.NoError(t, err, "unexpected error received")

			testMsg := "time test"
			l.Info(testMsg)
			require.Len(t, w.arr, 1, "unexpected amount of logs received")

			var entry logEntry
			err = json.Unmarshal(w.arr[0], &entry)
			require.NoError(t, err, "failed to unmarshal log entry")

			require.Equal(t, testMsg, entry.Msg, "unexpected log message")

			_, err = time.Parse(tC.expectedFmt, entry.Time)
			require.NoError(t, err, "unexpected time format")
		})
	}
}

func TestLogger_additionalFields(t *testing.T) {
	w := newCustomWriter()

	testCases := []struct {
		name     string
		msg      string
		fields   []any // As a key-value pairs.
		expected map[string]any
	}{
		{
			name:   "single field",
			msg:    "user login",
			fields: []any{"user_id", 123},
			expected: map[string]any{
				"msg":     "user login",
				"user_id": float64(123),
			},
		},
		{
			name:   "multiple fields",
			msg:    "request processed",
			fields: []any{"method", "GET", "path", "/api", "status", 200},
			expected: map[string]any{
				"msg":    "request processed",
				"method": "GET",
				"path":   "/api",
				"status": float64(200),
			},
		},
		{
			name:   "nested fields",
			msg:    "system event",
			fields: []any{"details", map[string]any{"service": "auth", "code": "E100"}},
			expected: map[string]any{
				"msg": "system event",
				"details": map[string]any{
					"service": "auth",
					"code":    "E100",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w.CleanUp()
			l, err := NewLogger("json", "debug", time.UnixDate, w)
			require.NoError(t, err)

			l.Info(tc.msg, tc.fields...)

			require.Len(t, w.arr, 1, "unexpected amount of logs received")

			// JSON decoding.
			var logData map[string]any
			err = json.Unmarshal(w.arr[0], &logData)
			require.NoError(t, err, "got error, expected nil")

			// Checking necessary fields.
			require.Equal(t, "INFO", logData["level"])
			require.Equal(t, tc.msg, logData["msg"])
			_, err = time.Parse(time.UnixDate, logData["time"].(string))
			require.NoError(t, err, "got error, expected nil")

			// Checking additional fields.
			for key, expectedValue := range tc.expected {
				if key == "msg" {
					continue
				}
				actualValue := logData[key]
				require.Equal(t, expectedValue, actualValue,
					"invalid value for %s", key)
			}
		})
	}
}

func TestLogger_With(t *testing.T) {
	w := newCustomWriter()
	l, err := NewLogger("json", "debug", time.UnixDate, w)
	require.NoError(t, err)

	loggerWithFields := l.With("user_id", 123, "service", "auth")

	testMsg := "operation completed"
	loggerWithFields.Info(testMsg)

	require.Len(t, w.arr, 1)

	var logData struct {
		Level   string `json:"level"`
		Msg     string `json:"msg"`
		UserID  int    `json:"user_id"` //nolint:tagliatelle
		Service string `json:"service"`
	}

	err = json.Unmarshal(w.arr[0], &logData)
	require.NoError(t, err)

	require.Equal(t, "INFO", logData.Level)
	require.Equal(t, testMsg, logData.Msg)
	require.Equal(t, 123, logData.UserID)
	require.Equal(t, "auth", logData.Service)
}
