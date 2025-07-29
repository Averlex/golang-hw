//go:build integration
// +build integration

// // Package integration implements integration tests for the project services.
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
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	// Local URL to the calendar service within the docker network.
	calendarServiceBaseURL = "http://calendar-test:9888/v1"

	defaultTimeFormat = time.RFC3339

	defaultTitle       = "Test event"
	defaultDatetime    = "01.01.2025 12:00:00.000"
	defaultDuration    = "3600s"
	defaultDescription = "Test description"
	defaultUserID      = "user123"
	alternativeUserID  = "user456"
	defaultRemindIn    = "1800s"

	userSingleEventID            = "event_single_user"
	alternativeUserSingleEventID = "event_single_user_alternative"
	userMultipleWeeklyID         = "event_multiple_weekly"
	userMultipleMonthlyID        = "event_multiple_monthly"

	nonExistingEventID = "01234567-8901-2345-6789-012345678901"
)

// testEventsData represents a set of test events designed to cover specific scenarios:
// 1. A single event for one user.
// 2. Multiple events for one user within a week.
// 3. Multiple events for one user within a month, spanning different weeks.
// All dates are hardcoded for reproducibility.
// The datetime format conforms to RFC3339 as indicated in the swagger spec (format: date-time).
var testEventsData []EventData = []EventData{
	// Scenario: One user, one event.
	{
		Title:       "Single User Event",
		Datetime:    "2035-08-17T10:00:00Z", // Friday
		Duration:    defaultDuration,        // "3600s"
		Description: "This user has only one event.",
		UserID:      userSingleEventID, // "event_single_user"
		RemindIn:    defaultRemindIn,   // "1800s"
	},
	// Scenario: Another user, one event, same day.
	{
		Title:       "Alternative User Event",
		Datetime:    "2035-08-17T12:00:00Z", // Friday
		Duration:    defaultDuration,        // "3600s"
		Description: "This user has only one event too.",
		UserID:      alternativeUserID, // "event_single_user"
		RemindIn:    defaultRemindIn,   // "1800s"
	},
	// Scenario: One user, multiple events in a week (week of 2035-08-18).
	{
		Title:       "User Weekly Event 1",
		Datetime:    "2035-08-18T09:00:00Z", // Saturday
		Duration:    defaultDuration,        // "3600s"
		Description: "First event of the week for this user.",
		UserID:      userMultipleWeeklyID, // "event_multiple_weekly"
		RemindIn:    defaultRemindIn,      // "1800s"
	},
	{
		Title:       "User Weekly Event 2",
		Datetime:    "2035-08-20T14:30:00Z", // Monday
		Duration:    defaultDuration,        // "3600s"
		Description: "Second event of the week for this user.",
		UserID:      userMultipleWeeklyID, // "event_multiple_weekly"
		RemindIn:    defaultRemindIn,      // "1800s"
	},
	{
		Title:       "User Weekly Event 3",
		Datetime:    "2035-08-22T11:00:00Z", // Wednesday
		Duration:    defaultDuration,        // "3600s"
		Description: "Third event of the week for this user.",
		UserID:      userMultipleWeeklyID, // "event_multiple_weekly"
		RemindIn:    defaultRemindIn,      // "1800s"
	},
	// Scenario: One user, multiple events in a month, different weeks (August 2035).
	// Week 1: July 30 - August 5, 2035.
	{
		Title:       "User Monthly Event Week 1",
		Datetime:    "2035-08-05T15:00:00Z", // Sunday
		Duration:    defaultDuration,        // "3600s"
		Description: "Event in the first week of the month.",
		UserID:      userMultipleMonthlyID, // "event_multiple_monthly"
		RemindIn:    defaultRemindIn,       // "1800s"
	},
	// Week 3: August 13 - August 19, 2035.
	{
		Title:       "User Monthly Event Week 2",
		Datetime:    "2035-08-13T10:30:00Z", // Monday
		Duration:    defaultDuration,        // "3600s"
		Description: "Event in the second week of the month.",
		UserID:      userMultipleMonthlyID, // "event_multiple_monthly"
		RemindIn:    defaultRemindIn,       // "1800s"
	},
	// Week 4: August 20 - August 26, 2035.
	{
		Title:       "User Monthly Event Week 3",
		Datetime:    "2035-08-21T13:45:00Z", // Tuesday
		Duration:    defaultDuration,        // "3600s"
		Description: "Event in the third week of the month.",
		UserID:      userMultipleMonthlyID, // "event_multiple_monthly"
		RemindIn:    defaultRemindIn,       // "1800s"
	},
	// Week 5: August 27 - September 02, 2035.
	{
		Title:       "User Monthly Event Week 4",
		Datetime:    "2035-08-27T16:00:00Z", // Monday
		Duration:    defaultDuration,        // "3600s"
		Description: "Event in the fourth week of the month.",
		UserID:      userMultipleMonthlyID, // "event_multiple_monthly"
		RemindIn:    defaultRemindIn,       // "1800s"
	},
}

// EventData represents the data for a calendar event considering client side format.
type EventData struct {
	Title       string `json:"title"`
	Datetime    string `json:"datetime"`
	Duration    string `json:"duration"` // Duration string, e.g., "1h30m".
	Description string `json:"description"`
	UserID      string `json:"user_id"`
	RemindIn    string `json:"remind_in"` // Duration string.
}

// ResponseEventData represents the data for a calendar event considering client side format.
type ResponseEventData struct {
	Title       string `json:"title"`
	Datetime    string `json:"datetime"`
	Duration    string `json:"duration"` // Duration string, e.g., "1h30m".
	Description string `json:"description"`
	UserID      string `json:"userId"`
	RemindIn    string `json:"remindIn"` // Duration string.
}

type CalendarIntegrationSuite struct {
	suite.Suite
	client        *http.Client
	createdEvents []string
}

func (s *CalendarIntegrationSuite) SetupSuite() {
	s.client = &http.Client{Timeout: 10 * time.Second}
}

func (s *CalendarIntegrationSuite) TearDownSuite() {
	s.client = nil
}

func (s *CalendarIntegrationSuite) SetupTest() {
	s.createdEvents = make([]string, 0)
	for _, event := range testEventsData {
		s.createTestEvent(event, http.StatusOK, false)
	}
}

func (s *CalendarIntegrationSuite) TearDownTest() {
	for _, id := range s.createdEvents {
		s.deleteTestEvent(id, http.StatusOK, false)
	}
	s.createdEvents = nil
}

// sendRequest is a helper for HTTP requests sending.
func (s *CalendarIntegrationSuite) sendRequest(method, url string, byteData []byte, expectedStatus int) []byte {
	s.T().Helper()
	var respBody []byte

	// Create the HTTP request.
	req, err := http.NewRequestWithContext(context.Background(), method, url, bytes.NewBuffer(byteData))
	s.Require().NoError(err, "creating create request")
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

	respBody, err = io.ReadAll(resp.Body)
	s.Require().NoError(err, "reading response body")
	s.Require().Equal(expectedStatus, resp.StatusCode, "status code mismatch with body: %s", string(respBody))

	return respBody
}

// validateResponseEvent is a helper which validates the response event data.
func (s *CalendarIntegrationSuite) validateResponseEvent(respBody []byte, eventData EventData, id *string, isCreated bool) {
	s.T().Helper()
	type EventResponse struct {
		ID   string            `json:"id"`
		Data ResponseEventData `json:"data"`
	}
	resp := EventResponse{}
	err := json.Unmarshal(respBody, &resp)
	s.Require().NoError(err, "unmarshalling response body")

	// Validate the returned event data matches what was sent (where applicable).
	if id != nil {
		s.Require().Equal(*id, resp.ID, "event ID mismatch")
	} else {
		s.Require().NotEmpty(resp.ID, "expected non-empty event ID, but got an empty one")
	}

	// Filling the slice for further clean up. Using only for created objects to prevent
	// triggering on get/update methods.
	if isCreated {
		s.createdEvents = append(s.createdEvents, resp.ID)
	}

	s.Require().Equal(eventData.Title, resp.Data.Title, "title mismatch")
	s.Require().Equal(eventData.Description, resp.Data.Description, "description mismatch")
	s.Require().Equal(eventData.UserID, resp.Data.UserID, "user ID mismatch")
	s.Require().NotEmpty(resp.Data.Datetime, "expected non-empty datetime, but got an empty one")
	expectedDuration, _ := time.ParseDuration(eventData.Duration)
	gotDuration, err := time.ParseDuration(resp.Data.Duration)
	s.Require().NoError(err, "parsing duration")
	s.Require().Equal(expectedDuration, gotDuration, "duration mismatch")
	if eventData.RemindIn != "" && eventData.RemindIn != "0s" {
		expectedDuration, _ := time.ParseDuration(eventData.RemindIn)
		gotDuration, err := time.ParseDuration(resp.Data.RemindIn)
		s.Require().NoError(err, "parsing remind_in")
		s.Require().Equal(expectedDuration, gotDuration, "remind_in mismatch")
	} else {
		s.Require().Truef(
			resp.Data.RemindIn == "" || resp.Data.RemindIn == "0s",
			"expected empty remind_in, but got %s", resp.Data.RemindIn,
		)
	}
}

// validateResponseEvents is a helper which validates the response event data for batch get methods.
func (s *CalendarIntegrationSuite) validateResponseEvents(respBody []byte, eventsData []EventData, ids []*string) {
	s.T().Helper()
	type EventResponse struct {
		ID   string            `json:"id"`
		Data ResponseEventData `json:"data"`
	}
	resp := []EventResponse{}
	err := json.Unmarshal(respBody, &resp)
	s.Require().NoError(err, "unmarshalling response body")
	s.Require().NotEmpty(resp, "expected non-empty response body, but got an empty one")

	s.Require().Lenf(resp, len(eventsData), "number of events mismatch: expected %v, got %v", len(eventsData), len(resp))

	for i, eventData := range eventsData {
		// Validate the returned event data matches what was sent (where applicable).
		if ids[i] != nil {
			// No guarantee that the returned order will be preserved as the source's one.
			hasMatch := slices.Contains[[]string, string](s.createdEvents, *ids[i])
			s.Require().Truef(hasMatch, "no match found for event ID: %s", *ids[i])
		} else {
			s.Require().NotEmpty(resp[i].ID, "expected non-empty event ID, but got an empty one")
		}

		// Initial events order may also be different.
		pos := -1
		for j, event := range resp {
			if event.Data.Title == eventData.Title {
				pos = j
				break
			}
		}
		if pos < 0 {
			s.Require().FailNow("unexpected event received")
		}

		s.Require().Equal(eventData.Title, resp[pos].Data.Title, "title mismatch")
		s.Require().Equal(eventData.Description, resp[pos].Data.Description, "description mismatch")
		s.Require().Equal(eventData.UserID, resp[pos].Data.UserID, "user ID mismatch")
		s.Require().NotEmpty(resp[pos].Data.Datetime, "expected non-empty datetime, but got an empty one")
		expectedDuration, _ := time.ParseDuration(eventData.Duration)
		gotDuration, err := time.ParseDuration(resp[pos].Data.Duration)
		s.Require().NoError(err, "parsing duration")
		s.Require().Equal(expectedDuration, gotDuration, "duration mismatch")
		if eventData.RemindIn != "" && eventData.RemindIn != "0s" {
			expectedDuration, _ := time.ParseDuration(eventData.RemindIn)
			gotDuration, err := time.ParseDuration(resp[pos].Data.RemindIn)
			s.Require().NoError(err, "parsing remind_in")
			s.Require().Equal(expectedDuration, gotDuration, "remind_in mismatch")
		} else {
			s.Require().Truef(
				resp[pos].Data.RemindIn == "" || resp[pos].Data.RemindIn == "0s",
				"expected empty remind_in, but got %s", resp[pos].Data.RemindIn,
			)
		}
	}
}

// createTestEvent is a helpler that creates a new event with the given data.
//
// Method fills the inner s.createdEvents slice if the event was created successfully.
func (s *CalendarIntegrationSuite) createTestEvent(eventData EventData, expectedStatus int, expectError bool) {
	s.T().Helper()
	eventDataBytes, err := json.Marshal(eventData)
	s.Require().NoError(err, "marshalling event data")

	// Send the HTTP request.
	url := fmt.Sprintf("%s/events", calendarServiceBaseURL)
	respBody := s.sendRequest(http.MethodPost, url, eventDataBytes, expectedStatus)

	// No following checks are needed.
	if expectError {
		return
	}

	s.validateResponseEvent(respBody, eventData, nil, true)
}

// deleteTestEvent is a helper which deletes the event with the given ID.
//
// Method updates the inner s.createdEvents slice.
func (s *CalendarIntegrationSuite) deleteTestEvent(id string, expectedStatus int, expectError bool) {
	s.T().Helper()
	requestData := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	requestDataBytes, err := json.Marshal(requestData)
	s.Require().NoError(err, "marshalling event data")

	// Create the HTTP request.
	url := fmt.Sprintf("%s/events/%s", calendarServiceBaseURL, id)
	respBody := s.sendRequest(http.MethodDelete, url, requestDataBytes, expectedStatus)

	// No following checks are needed.
	if expectError {
		s.Require().NotEmpty(respBody, "expected non-empty response body, but got an empty one")
		return
	}
}

// getTestEvent is a helper which implements default GET method for testing.
func (s *CalendarIntegrationSuite) getTestEvent(id string, expectedStatus int, expectError bool, dataToCompare *EventData) {
	s.T().Helper()
	requestData := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	requestDataBytes, err := json.Marshal(requestData)
	s.Require().NoError(err, "marshalling event data")

	// Create the HTTP request.
	url := fmt.Sprintf("%s/events/%s", calendarServiceBaseURL, id)
	respBody := s.sendRequest(http.MethodGet, url, requestDataBytes, expectedStatus)

	// No following checks are needed.
	if expectError || dataToCompare == nil {
		return
	}

	eventData := *dataToCompare
	s.validateResponseEvent(respBody, eventData, &id, false)
}

// TestCreateEvent tests the POST /events endpoint.
//
// Response data is not checked for datetime equality to avoid time format dependencies.
func (s *CalendarIntegrationSuite) TestCreateEvent() {
	validFutureTime := time.Now().Add(24 * time.Hour)
	invalidTime := "not-a-valid-date-time"

	testCases := []struct {
		name           string
		eventData      EventData
		expectedStatus int
		expectError    bool // If true, we only check for non-2xx status or an error condition, not specific content.
	}{
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
				RemindIn:    "0s",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "valid_create/datetime_overlaps_but_another_user",
			eventData: EventData{
				Title:       defaultTitle,
				Datetime:    validFutureTime.Format(defaultTimeFormat),
				Duration:    defaultDuration,
				Description: defaultDescription,
				UserID:      alternativeUserID,
				RemindIn:    defaultRemindIn,
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
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
	}

	ids := make([]string, 0)
	for _, tC := range testCases {
		s.Run(tC.name, func() {
			prevLen := len(s.createdEvents)
			s.createTestEvent(tC.eventData, tC.expectedStatus, tC.expectError)
			// Check the service state only if the new addition was successful.
			if len(s.createdEvents) > prevLen {
				ids = append(ids, s.createdEvents[len(s.createdEvents)-1])
				s.getTestEvent(ids[len(ids)-1], tC.expectedStatus, tC.expectError, &tC.eventData)
			}
		})
	}
}

// TestGetEvent tests the GET /events/{id} endpoint.
func (s *CalendarIntegrationSuite) TestGetEvent() {
	testCases := []struct {
		name           string
		id             string
		expectedStatus int
		expectError    bool // If true, we only check for non-2xx status or an error condition, not specific content.
		dataToCompare  *EventData
	}{
		{
			name:           "valid_get",
			id:             s.createdEvents[0],
			expectedStatus: http.StatusOK,
			expectError:    false,
			dataToCompare: &EventData{
				Title:       testEventsData[0].Title,
				Datetime:    testEventsData[0].Datetime,
				Duration:    testEventsData[0].Duration,
				Description: testEventsData[0].Description,
				UserID:      testEventsData[0].UserID,
				RemindIn:    testEventsData[0].RemindIn,
			},
		},
		{
			name:           "invalid_id",
			id:             "invalid_id",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			dataToCompare:  nil,
		},
		{
			name:           "id_does_not_exist",
			id:             nonExistingEventID,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			dataToCompare:  nil,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			s.getTestEvent(tC.id, tC.expectedStatus, tC.expectError, tC.dataToCompare)
		})
	}
}

// TestDeleteEvent tests the DELETE /events/{id} endpoint.
func (s *CalendarIntegrationSuite) TestDeleteEvent() {
	testCases := []struct {
		name           string
		id             string
		expectedStatus int
		expectError    bool // If true, we only check for non-2xx status or an error condition, not specific content.
	}{
		{
			name:           "valid_delete",
			id:             s.createdEvents[0],
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid_id",
			id:             "invalid_id",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "id_does_not_exist",
			id:             nonExistingEventID,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			s.deleteTestEvent(tC.id, tC.expectedStatus, tC.expectError)
			if !tC.expectError {
				s.createdEvents = append(s.createdEvents[:0], s.createdEvents[1:]...)
			}
		})
	}
}

// TestUpdateEvent tests the PUT /events/{id} endpoint.
func (s *CalendarIntegrationSuite) TestUpdateEvent() {
	testCases := []struct {
		name           string
		id             string
		data           EventData
		expectedStatus int
		expectError    bool // If true, we only check for non-2xx status or an error condition, not specific content.
	}{
		{
			name: "valid_update/no_overlaps",
			id:   s.createdEvents[0],
			data: func() EventData {
				event := testEventsData[0]
				event.Datetime = "2045-08-15T10:00:00Z"
				event.Title = "Updated title"
				event.Description = ""
				event.RemindIn = "0s"
				return event
			}(),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "valid_update/datetime_overlaps_with_itself",
			id:             s.createdEvents[2],
			data:           testEventsData[2],
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid_update/invalid_request",
			id:             "invalid_id",
			data:           testEventsData[2],
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid_update/invalid_data",
			id:             s.createdEvents[2],
			data:           EventData{},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid_update/not_found",
			id:             nonExistingEventID,
			data:           testEventsData[2],
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name: "invalid_update/date_busy",
			id:   s.createdEvents[2],
			data: func() EventData {
				event := testEventsData[2]
				event.Datetime = testEventsData[3].Datetime
				return event
			}(),
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
		{
			name: "invalid_update/modifying_another_user_event",
			id:   s.createdEvents[0],
			data: func() EventData {
				event := testEventsData[2]
				return event
			}(),
			expectedStatus: http.StatusForbidden,
			expectError:    true,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			requestDataBytes, err := json.Marshal(tC.data)
			s.Require().NoError(err, "marshalling event data")

			// Send the HTTP request.
			url := fmt.Sprintf("%s/events/%s", calendarServiceBaseURL, tC.id)
			respBody := s.sendRequest(http.MethodPut, url, requestDataBytes, tC.expectedStatus)

			// No following checks are needed.
			if tC.expectError {
				return
			}

			s.validateResponseEvent(respBody, tC.data, &tC.id, false)

			// Check that the event was really updated.
			s.getTestEvent(tC.id, http.StatusOK, tC.expectError, &tC.data)
		})
	}
}

// TestGetEventsForDay tests the GET /events/day?date={date}&userId={user_id} endpoint.
func (s *CalendarIntegrationSuite) TestGetEventsForDay() {
	testCases := []struct {
		name           string
		ids            []*string
		datetime       string
		userID         string
		expectedStatus int
		expectError    bool // If true, we only check for non-2xx status or an error condition, not specific content.
		dataToCompare  []EventData
	}{
		{
			name:           "valid_get/one_event_one_user",
			ids:            []*string{&s.createdEvents[0]},
			datetime:       "2035-08-17T01:00:00Z", // Shifted testEventsData[0].Datetime value
			userID:         testEventsData[0].UserID,
			expectedStatus: http.StatusOK,
			expectError:    false,
			dataToCompare: func() []EventData {
				event := testEventsData[0]
				return []EventData{event}
			}(),
		},
		{
			name:           "valid_get/several_events_no_user",
			ids:            []*string{&s.createdEvents[0], &s.createdEvents[1]},
			datetime:       testEventsData[0].Datetime,
			userID:         "",
			expectedStatus: http.StatusOK,
			expectError:    false,
			dataToCompare: func() []EventData {
				res := make([]EventData, 0)
				event1 := testEventsData[0]
				event2 := testEventsData[1]
				res = append(res, event1)
				res = append(res, event2)
				return res
			}(),
		},
		{
			name:           "invalid_get/invalid_request",
			datetime:       "",
			userID:         "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			dataToCompare:  nil,
		},
		{
			name:           "invalid_get/events_not_found",
			datetime:       "2055-08-17T04:00:00Z",
			userID:         "",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			dataToCompare:  nil,
		},
		{
			name:           "invalid_get/user_not_found",
			datetime:       "2035-08-17T04:00:00Z", // Shifted testEventsData[0].Datetime value
			userID:         "fictional_user_id",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			dataToCompare:  nil,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			// Send the HTTP request.
			var userQueryParam string
			if tC.userID != "" {
				userQueryParam = fmt.Sprintf("&userId=%s", tC.userID)
			}
			url := fmt.Sprintf("%s/events/day?date=%s%s", calendarServiceBaseURL, tC.datetime, userQueryParam)
			respBody := s.sendRequest(http.MethodGet, url, nil, tC.expectedStatus)

			// No following checks are needed.
			if tC.expectError || len(tC.dataToCompare) == 0 {
				return
			}

			s.validateResponseEvents(respBody, tC.dataToCompare, tC.ids)
		})
	}
}

// TestGetEventsForWeek tests the GET /events/week?date={date}&userId={user_id} endpoint.
func (s *CalendarIntegrationSuite) TestGetEventsForWeek() {
	testCases := []struct {
		name           string
		ids            []*string
		datetime       string
		userID         string
		expectedStatus int
		expectError    bool // If true, we only check for non-2xx status or an error condition, not specific content.
		dataToCompare  []EventData
	}{
		{
			name:           "valid_get/one_event_one_user",
			ids:            []*string{&s.createdEvents[0]},
			datetime:       "2035-08-17T01:00:00Z", // Shifted testEventsData[0].Datetime value
			userID:         testEventsData[0].UserID,
			expectedStatus: http.StatusOK,
			expectError:    false,
			dataToCompare: func() []EventData {
				event := testEventsData[0]
				return []EventData{event}
			}(),
		},
		{
			name:           "valid_get/several_events_no_user",
			ids:            []*string{&s.createdEvents[0], &s.createdEvents[1], &s.createdEvents[2], &s.createdEvents[6]},
			datetime:       "2035-08-14T01:23:45Z", // Shifted testEventsData[0].Datetime value (Friday -> Tuesday)
			userID:         "",
			expectedStatus: http.StatusOK,
			expectError:    false,
			dataToCompare:  []EventData{testEventsData[0], testEventsData[1], testEventsData[2], testEventsData[6]},
		},
		{
			name:           "valid_get/several_events_one_user",
			ids:            []*string{&s.createdEvents[3], &s.createdEvents[4]},
			datetime:       "2035-08-26T23:59:59Z", // Shifted testEventsData[3].Datetime value (Monday -> Sunday)
			userID:         testEventsData[3].UserID,
			expectedStatus: http.StatusOK,
			expectError:    false,
			dataToCompare:  []EventData{testEventsData[3], testEventsData[4]},
		},
		{
			name:           "invalid_get/invalid_request",
			datetime:       "",
			userID:         "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			dataToCompare:  nil,
		},
		{
			name:           "invalid_get/events_not_found",
			datetime:       "2055-08-17T04:00:00Z",
			userID:         "",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			dataToCompare:  nil,
		},
		{
			name:           "invalid_get/user_not_found",
			datetime:       "2035-08-17T04:00:00Z", // Shifted testEventsData[0].Datetime value
			userID:         "fictional_user_id",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			dataToCompare:  nil,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			// Send the HTTP request.
			var userQueryParam string
			if tC.userID != "" {
				userQueryParam = fmt.Sprintf("&userId=%s", tC.userID)
			}
			url := fmt.Sprintf("%s/events/week?date=%s%s", calendarServiceBaseURL, tC.datetime, userQueryParam)
			respBody := s.sendRequest(http.MethodGet, url, nil, tC.expectedStatus)

			// No following checks are needed.
			if tC.expectError || len(tC.dataToCompare) == 0 {
				return
			}

			s.validateResponseEvents(respBody, tC.dataToCompare, tC.ids)
		})
	}
}

// TestCalendarIntegrationSuite runs the suite.
func TestCalendarIntegrationSuite(t *testing.T) {
	suite.Run(t, new(CalendarIntegrationSuite))
}
