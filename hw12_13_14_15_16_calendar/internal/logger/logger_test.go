package logger_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/logger" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors"      //nolint:depguard,nolintlint
	"github.com/stretchr/testify/suite"                                      //nolint:depguard,nolintlint
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

type LoggerTestSuite struct {
	suite.Suite
	writer *customWriter
}

func (s *LoggerTestSuite) SetupTest() {
	s.writer = newCustomWriter()
}

func (s *LoggerTestSuite) TearDownTest() {
	s.writer.CleanUp()
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

func (s *LoggerTestSuite) TestDefaults() {
	s.Run("set defaults", func() {
		s.writer.CleanUp()
		l, err := logger.NewLogger(logger.SetDefaults(), logger.WithWriter(s.writer))
		s.Require().NoError(err, "got error, expected nil")

		l.Info(context.Background(), "test")
		l.Error(context.Background(), "test again")
		s.Require().Len(s.writer.arr, 1, "unexpected amount of logs received")

		entry, err := decodeJSON(s.writer.arr[0])
		s.Require().NoError(err, "failed to unmarshal log entry")
		s.Require().Equal("test again", entry.Msg, "unexpected log message")
		s.Require().Equal("error", strings.ToLower(entry.Level), "unexpected log level")

		logTime, err := time.ParseInLocation("02.01.2006 15:04:05.000", entry.Time, time.Local)
		s.Require().NoError(err, "got error, expected nil")
		s.Require().InDelta(time.Now().UnixMilli(), logTime.UnixMilli(), float64(500),
			"measured time doesn't match the logged one")
	})

	s.Run("nil writer", func() {
		s.writer.CleanUp()
		_, err := logger.NewLogger(logger.WithWriter(nil))
		s.Require().Error(err, "expected error for nil writer")
		s.Require().ErrorIs(err, errors.ErrLoggerInitFailed)
	})
}

func (s *LoggerTestSuite) TestLogLevel() {
	callOrder := []string{"debug", "info", "warn", "error"}
	testCases := []struct {
		name         string
		level        string
		msg          string
		expectedSize int
		expectError  bool
	}{
		{"debug", "debug", "debug-test", 4, false},
		{"info", "info", "info-test", 3, false},
		{"warn", "warn", "warn-test", 2, false},
		{"error", "error", "error-test", 1, false},
		{"case insensitivity", "dEbUg", "debug-test", 4, false},
		{"empty level", "", "empty-test", 1, false},
		{"unknown", "unknown", "unknown-test", 0, true},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			s.writer.CleanUp()
			opts := []logger.Option{
				logger.WithConfig(map[string]any{
					"level":         tC.level,
					"time_template": time.UnixDate,
					"log_stream":    "stdout",
				}),
				logger.WithWriter(s.writer),
			}
			l, err := logger.NewLogger(opts...)
			if tC.expectError {
				s.Require().Error(err, "got nil, expected error")
				s.Require().ErrorIs(err, errors.ErrCorruptedConfig)
				return
			}
			s.Require().NoError(err, "got error, expected nil")

			l.Debug(context.Background(), tC.msg)
			l.Info(context.Background(), tC.msg)
			l.Warn(context.Background(), tC.msg)
			l.Error(context.Background(), tC.msg)

			s.Require().Len(s.writer.arr, tC.expectedSize, "unexpected amount of logs received")
			for i, data := range s.writer.arr {
				var entry logEntry
				err := json.Unmarshal(data, &entry)
				s.Require().NoError(err, "failed to unmarshal log entry")
				s.Require().Equal(tC.msg, entry.Msg, "unexpected log message")
				s.Require().Equal(callOrder[len(callOrder)-tC.expectedSize:][i], strings.ToLower(entry.Level),
					"unexpected log level")
			}
		})
	}
}

func (s *LoggerTestSuite) TestLogType() {
	testCases := []struct {
		name                   string
		logType                string
		msg                    string
		expectConstructorError bool
		expectDecodingError    bool
	}{
		{"json", "json", "json-test", false, false},
		{"text", "text", "text-test", false, true},
		{"empty log type", "", "empty-test", false, false},
		{"unknown", "unknown", "unknown-test", true, false},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			s.writer.CleanUp()
			opts := []logger.Option{
				logger.WithConfig(map[string]any{
					"format":        tC.logType,
					"level":         "info",
					"time_template": time.UnixDate,
					"log_stream":    "stdout",
				}),
				logger.WithWriter(s.writer),
			}
			l, err := logger.NewLogger(opts...)
			if tC.expectConstructorError {
				s.Require().Error(err, "got nil, expected error")
				s.Require().ErrorIs(err, errors.ErrCorruptedConfig)
				return
			}
			s.Require().NoError(err, "got error, expected nil")

			l.Info(context.Background(), tC.msg)
			s.Require().Len(s.writer.arr, 1, "unexpected amount of logs received")

			_, err = decodeJSON(s.writer.arr[0])
			if tC.expectDecodingError {
				s.Require().Error(err, "got nil, expected error")
				return
			}
			s.Require().NoError(err, "got error, expected nil")
		})
	}
}

func (s *LoggerTestSuite) TestTimeTemplate() {
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
		s.Run(tC.name, func() {
			s.writer.CleanUp()
			opts := []logger.Option{
				logger.WithConfig(map[string]any{
					"level":         "info",
					"format":        "json",
					"time_template": tC.template,
					"log_stream":    "stdout",
				}),
				logger.WithWriter(s.writer),
			}
			l, err := logger.NewLogger(opts...)
			if tC.expectError {
				s.Require().Error(err, "got nil, expected error")
				s.Require().ErrorIs(err, errors.ErrCorruptedConfig)
				return
			}
			s.Require().NoError(err, "unexpected error received")

			testMsg := "time test"
			l.Info(context.Background(), testMsg)
			s.Require().Len(s.writer.arr, 1, "unexpected amount of logs received")

			var entry logEntry
			err = json.Unmarshal(s.writer.arr[0], &entry)
			s.Require().NoError(err, "failed to unmarshal log entry")
			s.Require().Equal(testMsg, entry.Msg, "unexpected log message")

			_, err = time.Parse(tC.expectedFmt, entry.Time)
			s.Require().NoError(err, "unexpected time format")
		})
	}
}

func (s *LoggerTestSuite) TestAdditionalFields() {
	testCases := []struct {
		name     string
		msg      string
		fields   []any
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
		{
			name:     "no fields",
			msg:      "no fields",
			fields:   []any{},
			expected: map[string]any{"msg": "no fields"},
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			s.writer.CleanUp()
			l, err := logger.NewLogger(
				logger.WithConfig(map[string]any{
					"level":         "debug",
					"time_template": time.UnixDate,
					"log_stream":    "stdout",
				}),
				logger.WithWriter(s.writer),
			)
			s.Require().NoError(err, "got error, expected nil")

			l.Info(context.Background(), tC.msg, tC.fields...)
			s.Require().Len(s.writer.arr, 1, "unexpected amount of logs received")

			var logData map[string]any
			err = json.Unmarshal(s.writer.arr[0], &logData)
			s.Require().NoError(err, "failed to unmarshal log entry")

			s.Require().Equal("INFO", logData["level"], "unexpected log level")
			s.Require().Equal(tC.msg, logData["msg"], "unexpected log message")
			_, err = time.Parse(time.UnixDate, logData["time"].(string))
			s.Require().NoError(err, "unexpected time format")

			for key, expectedValue := range tC.expected {
				if key == "msg" {
					continue
				}
				actualValue := logData[key]
				s.Require().Equal(expectedValue, actualValue, "invalid value for %s", key)
			}
		})
	}
}

func (s *LoggerTestSuite) TestWith() {
	testCases := []struct {
		name     string
		fields   []any
		msg      string
		expected map[string]any
	}{
		{
			name:   "single field",
			fields: []any{"user_id", 123},
			msg:    "operation completed",
			expected: map[string]any{
				"user_id": float64(123),
				"msg":     "operation completed",
			},
		},
		{
			name:   "multiple fields",
			fields: []any{"user_id", 123, "service", "auth"},
			msg:    "operation completed",
			expected: map[string]any{
				"user_id": float64(123),
				"service": "auth",
				"msg":     "operation completed",
			},
		},
		{
			name:   "chained with",
			fields: []any{"user_id", 123},
			msg:    "chained operation",
			expected: map[string]any{
				"user_id": float64(123),
				"service": "auth",
				"msg":     "chained operation",
			},
		},
		{
			name:     "empty fields",
			fields:   []any{},
			msg:      "no fields",
			expected: map[string]any{"msg": "no fields"},
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			s.writer.CleanUp()
			l, err := logger.NewLogger(
				logger.WithConfig(map[string]any{
					"level":         "debug",
					"time_template": time.UnixDate,
					"log_stream":    "stdout",
				}),
				logger.WithWriter(s.writer),
			)
			s.Require().NoError(err, "got error, expected nil")

			loggerWithFields := l
			if tC.name == "chained with" {
				loggerWithFields = l.With("service", "auth").With(tC.fields...)
			} else {
				loggerWithFields = l.With(tC.fields...)
			}

			loggerWithFields.Info(context.Background(), tC.msg)
			s.Require().Len(s.writer.arr, 1, "unexpected amount of logs received")

			var logData map[string]any
			err = json.Unmarshal(s.writer.arr[0], &logData)
			s.Require().NoError(err, "failed to unmarshal log entry")

			s.Require().Equal("INFO", logData["level"], "unexpected log level")
			s.Require().Equal(tC.msg, logData["msg"], "unexpected log message")
			_, err = time.Parse(time.UnixDate, logData["time"].(string))
			s.Require().NoError(err, "unexpected time format")

			for key, expectedValue := range tC.expected {
				if key == "msg" {
					continue
				}
				actualValue := logData[key]
				s.Require().Equal(expectedValue, actualValue, "invalid value for %s", key)
			}
		})
	}
}

func (s *LoggerTestSuite) TestInvalidConfigTypes() {
	testCases := []struct {
		name          string
		config        map[string]any
		expectedError error
	}{
		{
			name:          "invalid level type",
			config:        map[string]any{"level": 123, "time_template": time.UnixDate, "log_stream": "stdout"},
			expectedError: errors.ErrCorruptedConfig,
		},
		{
			name: "invalid log type",
			config: map[string]any{
				"format": 123, "level": "info",
				"time_template": time.UnixDate, "log_stream": "stdout",
			},
			expectedError: errors.ErrCorruptedConfig,
		},
		{
			name:          "invalid writer type",
			config:        map[string]any{"log_stream": 123, "level": "info", "time_template": time.UnixDate},
			expectedError: errors.ErrCorruptedConfig,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			s.writer.CleanUp()
			_, err := logger.NewLogger(
				logger.WithConfig(tC.config),
				logger.WithWriter(s.writer),
			)
			s.Require().Error(err, "got nil, expected error")
			s.Require().ErrorIs(err, tC.expectedError)
		})
	}
}
