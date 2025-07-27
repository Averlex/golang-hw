// Package integration implements integration tests for the project services.
//
//nolint:all
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	// Local URL to the calendar service within the docker network.
	calendarServiceBaseURL = "http://calendar:8888/v1"

	defaultTimeFormat = "02.01.2006 15:04:05.000"

	defaultTitle       = "Test event"
	defaultDatetime    = "01.01.2025 12:00:00.000"
	defaultDuration    = "3600s"
	defaultDescription = "Test description"
	defaultUserID      = "user123"
	defaultRemindIn    = "1800s"
)

// EventData represents the data for a calendar event considering client side format.
type EventData struct {
	Title       string `json:"title"`
	Datetime    string `json:"datetime"`
	Duration    string `json:"duration"` // Duration string, e.g., "1h30m".
	Description string `json:"description"`
	UserID      string `json:"user_id"`
	RemindIn    string `json:"remind_in"` // Duration string
}

type testCase struct {
	name           string
	eventData      EventData
	expectedStatus int
	expectError    bool // If true, we only check for non-2xx status or an error condition, not specific content.
}

type CalendarIntegrationSuite struct {
	suite.Suite
	client *http.Client
}

func (s *CalendarIntegrationSuite) SetupSuite() {
	s.client = &http.Client{Timeout: 10 * time.Second}
}

func (s *CalendarIntegrationSuite) TearDownSuite() {
	s.client = nil
}

// TestCreateEvent tests the POST /events endpoint.
//
// Response data is not checked for datetime equality to avoid time format dependencies.
func (s *CalendarIntegrationSuite) TestCreateEvent() {
	validFutureTime := time.Now().Add(24 * time.Hour)
	invalidTime := "not-a-valid-date-time"

	testCases := []testCase{
		{
			name: "valid_create/all_data",
			eventData: EventData{
				Title:       defaultTitle,
				Datetime:    validFutureTime.Format(defaultTimeFormat),
				Duration:    defaultDuration,
				Description: defaultDescription,
				UserID:      defaultUserID,
				RemindIn:    defaultRemindIn,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "valid_create/empty_description",
			eventData: EventData{
				Title:    defaultTitle,
				Datetime: validFutureTime.Add(24 * time.Hour).Format(defaultTimeFormat),
				Duration: defaultDuration,
				// Description is missing.
				UserID:   defaultUserID,
				RemindIn: defaultRemindIn,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "valid_create/empty_remind_in",
			eventData: EventData{
				Title:       defaultTitle,
				Datetime:    validFutureTime.Add(48 * time.Hour).Format(defaultTimeFormat),
				Duration:    defaultDuration,
				Description: defaultDescription,
				UserID:      defaultUserID,
				// RemindIn is missing.
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "invalid_data/empty_title",
			eventData: EventData{
				// Title is missing.
				Datetime:    validFutureTime.Add(72 * time.Hour).Format(defaultTimeFormat),
				Duration:    defaultDuration,
				Description: defaultDescription,
				UserID:      defaultUserID,
				RemindIn:    defaultRemindIn,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid_data/empty_user_id",
			eventData: EventData{
				Title:       defaultTitle,
				Datetime:    time.Now().Add(72 * time.Hour).Format(defaultTimeFormat),
				Duration:    defaultDuration,
				Description: defaultDescription,
				// UserID is missing.
				RemindIn: defaultRemindIn,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid_data/wrong_date_format",
			eventData: EventData{
				Title:       defaultTitle,
				Datetime:    invalidTime,
				Duration:    defaultDuration,
				Description: defaultDescription,
				UserID:      defaultUserID,
				RemindIn:    defaultRemindIn,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid_data/date_is_busy",
			eventData: EventData{
				Title:       defaultTitle,
				Datetime:    validFutureTime.Format(defaultTimeFormat), // The same as in the first case.
				Duration:    defaultDuration,
				Description: defaultDescription,
				UserID:      defaultUserID,
				RemindIn:    defaultRemindIn,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			// Marshal the event data to JSON.
			eventDataBytes, err := json.Marshal(tC.eventData)
			s.Require().NoError(err, "marshalling event data")

			// Create the HTTP request.
			url := fmt.Sprintf("%s/events", calendarServiceBaseURL)
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(eventDataBytes))
			s.Require().NoError(err, "creating request")
			req.Header.Set("Content-Type", "application/json")

			// Perform the request.
			resp, err := s.client.Do(req)
			s.Require().NoError(err, "HTTP request failed")
			defer func() {
				// Ensure the response body is closed to prevent resource leaks.
				if resp != nil && resp.Body != nil {
					_, _ = io.Copy(io.Discard, resp.Body) // Discard any remaining response body.
					resp.Body.Close()
				}
			}()

			s.Require().Equal(tC.expectedStatus, resp.StatusCode, "status code mismatch")

			// Read and parse the response body
			body, err := io.ReadAll(resp.Body)
			s.Require().NoError(err, "reading response body")

			type CreateEventResponse struct {
				ID   string    `json:"id"`
				Data EventData `json:"data"`
			}
			createResp := CreateEventResponse{}
			err = json.Unmarshal(body, &createResp)
			s.Require().NoError(err, "unmarshalling response body")

			// Validate the returned event data matches what was sent (where applicable).
			s.Require().NotEmpty(createResp.ID, "expected non-empty event ID, but got an empty one")
			s.Require().Equal(tC.eventData.Title, createResp.Data.Title, "title mismatch")
			s.Require().Equal(tC.eventData.Description, createResp.Data.Description, "description mismatch")
			s.Require().Equal(tC.eventData.UserID, createResp.Data.UserID, "user ID mismatch")
			s.Require().NotEmpty(createResp.Data.Datetime, "expected non-empty datetime, but got an empty one")
			expectedDuration, _ := time.ParseDuration(tC.eventData.Duration)
			gotDuration, err := time.ParseDuration(createResp.Data.Duration)
			s.Require().NoError(err, "parsing duration")
			s.Require().Equal(expectedDuration, gotDuration, "duration mismatch")
			if tC.eventData.RemindIn != "" {
				expectedDuration, _ := time.ParseDuration(tC.eventData.RemindIn)
				gotDuration, err := time.ParseDuration(createResp.Data.RemindIn)
				s.Require().NoError(err, "parsing remind_in")
				s.Require().Equal(expectedDuration, gotDuration, "remind_in mismatch")
			} else {
				s.Require().NotEmpty(createResp.Data.RemindIn, "expected non-empty remind_in, but got an empty one")
			}
		})
	}
}

// TestCalendarIntegrationSuite runs the suite.
func TestCalendarIntegrationSuite(t *testing.T) {
	suite.Run(t, new(CalendarIntegrationSuite))
}
