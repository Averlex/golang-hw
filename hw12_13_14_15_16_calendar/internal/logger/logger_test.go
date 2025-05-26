package logger

import (
	"encoding/json"
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
	t.Run("default logger", emptyParamsTest)
	t.Run("log level", logLevelTest)
	t.Run("log type", logTypeTest)
}

func emptyParamsTest(t *testing.T) {
	t.Helper()
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

// timeTemplate
// additional args
// incorrect args
// With()
