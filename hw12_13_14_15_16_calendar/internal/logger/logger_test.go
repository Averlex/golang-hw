package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type customWriter struct {
	arr [][]byte
}

func (w *customWriter) Write(data []byte) (int, error) {
	w.arr = append(w.arr, data)
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

func decodeJSON(data []byte) (map[string]interface{}, error) {
	var js map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&js)
	return js, err
}

func TestLogger(t *testing.T) {
	t.Run("default logger", emptyParams)
}

func emptyParams(t *testing.T) {
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

		js, err := decodeJSON(w.arr[0])
		require.NoError(t, err, "got error, expected nil")

		msg, ok := js["msg"].(string)
		require.True(t, ok, "unable cast msg to string")
		require.Equal(t, "test", msg)
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

		js, err := decodeJSON(w.arr[0])
		require.NoError(t, err, "got error, expected nil")

		lvl, ok := js["level"].(string)
		require.True(t, ok, "unable cast level to string")
		require.Equal(t, "error", strings.ToLower(lvl))
	})
	w.CleanUp()

	t.Run("empty time template", func(t *testing.T) {
		l, err := NewLogger("json", "info", "", w) // "02.01.2006 15:04:05.000" time template expected.
		require.NoError(t, err, "got error, expected nil")

		l.Info("test")
		require.Len(t, w.arr, 1)

		js, err := decodeJSON(w.arr[0])
		require.NoError(t, err, "got error, expected nil")

		msg, ok := js["time"].(string)
		require.True(t, ok, "unable cast time to string")

		logTime, err := time.ParseInLocation("02.01.2006 15:04:05.000", msg, time.Local)
		require.NoError(t, err, "got error, expected nil")
		require.InDelta(t, time.Now().UnixMilli(), logTime.UnixMilli(), float64(500),
			"measured time doesn't match the logged one") // Comparing with an error margin of 500ms.
	})
}
