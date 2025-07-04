package grpc_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1"                //nolint:depguard,nolintlint
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors"     //nolint:depguard,nolintlint
	calendarGRPC "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/server/grpc" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/server/grpc/mocks"        //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types"                    //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                                   //nolint:depguard,nolintlint
	"github.com/stretchr/testify/mock"                                                         //nolint:depguard,nolintlint
	"github.com/stretchr/testify/require"                                                      //nolint:depguard,nolintlint
	"github.com/stretchr/testify/suite"                                                        //nolint:depguard,nolintlint
	"google.golang.org/grpc"                                                                   //nolint:depguard,nolintlint
	"google.golang.org/grpc/codes"                                                             //nolint:depguard,nolintlint
	"google.golang.org/grpc/credentials/insecure"                                              //nolint:depguard,nolintlint
	"google.golang.org/grpc/status"                                                            //nolint:depguard,nolintlint
	"google.golang.org/grpc/test/bufconn"                                                      //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/durationpb"                                        //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/timestamppb"                                       //nolint:depguard,nolintlint
)

const basicUserID = "user1"

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}

type ServerSuite struct {
	suite.Suite
	grpcServer *grpc.Server
	listener   *bufconn.Listener
	app        *mocks.Application
	logger     *mocks.Logger
	client     pb.CalendarServiceClient
}

func (s *ServerSuite) loggerMocks(t *testing.T) {
	t.Helper()

	s.logger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	s.logger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	// This is actually a middleware call, and the middleware is currently disabled.
	s.logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
}

func (s *ServerSuite) setupTestServer(t *testing.T) {
	t.Helper()
	s.listener = bufconn.Listen(1024 * 1024)
	s.app = &mocks.Application{}
	s.logger = &mocks.Logger{}
	server, err := calendarGRPC.NewServer(s.logger, s.app, map[string]interface{}{
		"host":             "localhost",
		"port":             "0",
		"shutdown_timeout": time.Second,
	})
	require.NoError(t, err, "error on server creation")
	s.grpcServer = grpc.NewServer()
	pb.RegisterCalendarServiceServer(s.grpcServer, server)
	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			t.Logf("server stopped: %v", err)
		}
	}()
}

func (s *ServerSuite) setupTestClient(t *testing.T) {
	t.Helper()
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return s.listener.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err, "error on client setup")
	s.client = pb.NewCalendarServiceClient(conn)
}

func (s *ServerSuite) SetupTest() {
	s.setupTestServer(s.T())
	s.setupTestClient(s.T())
}

func (s *ServerSuite) TearDownTest() {
	s.grpcServer.GracefulStop()
	s.listener.Close()
	s.app.AssertExpectations(s.T())
	s.logger.AssertExpectations(s.T())
}

func (s *ServerSuite) TestGetEvent() {
	input := &types.Event{
		ID: uuid.New(),
		EventData: types.EventData{
			Title:       "Test Event",
			Datetime:    time.Now(),
			Duration:    time.Hour,
			Description: "Test Description",
			UserID:      basicUserID,
			RemindIn:    time.Minute * 30,
		},
	}
	expectedOutput := &pb.GetEventResponse{
		Event: &pb.Event{
			Id: input.ID.String(),
			Data: &pb.EventData{
				Title:       input.Title,
				Datetime:    timestamppb.New(input.Datetime),
				Duration:    durationpb.New(input.Duration),
				Description: input.Description,
				UserId:      input.UserID,
				RemindIn:    durationpb.New(input.RemindIn),
			},
		},
	}

	testCases := []struct {
		name         string
		req          *pb.GetEventRequest
		mockApp      func(*mocks.Application)
		want         *pb.GetEventResponse
		expectedCode codes.Code
	}{
		{
			name: "success",
			req:  &pb.GetEventRequest{Id: uuid.New().String()},
			mockApp: func(m *mocks.Application) {
				m.On("GetEvent", mock.Anything, mock.Anything).Return(input, nil).Once()
			},
			want:         expectedOutput,
			expectedCode: codes.OK,
		},
		{
			name: "invalid uuid",
			req:  &pb.GetEventRequest{Id: "invalid-uuid"},
			mockApp: func(_ *mocks.Application) {
				// No mock setup as parseUUID will fail before calling Application.
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "event not found",
			req:  &pb.GetEventRequest{Id: uuid.New().String()},
			mockApp: func(m *mocks.Application) {
				m.On("GetEvent", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrEventNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  &pb.GetEventRequest{Id: uuid.New().String()},
			mockApp: func(m *mocks.Application) {
				m.On("GetEvent", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrInconsistentState).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.mockApp(s.app)
			s.loggerMocks(s.T())
			resp, err := s.client.GetEvent(context.Background(), tC.req)

			if tC.expectedCode != codes.OK {
				s.Require().Error(err, "expected error, got nil")
				s.Require().Nil(resp, "expected nil response on error, got non-nil")
				s.Require().Equal(tC.expectedCode, status.Code(err), "unexpected error code")
				return
			}

			s.Require().NoError(err, "unexpected error on GetEvent")
			s.Require().NotNil(resp, "got nil response, expected non-nil")
			s.Require().Equal(tC.want.Event.Data.Title, resp.Event.Data.Title, "title mismatch")
			s.Require().Equal(tC.want.Event.Data.Description, resp.Event.Data.Description, "description mismatch")
			s.Require().Equal(tC.want.Event.Data.UserId, resp.Event.Data.UserId, "user ID mismatch")
			s.Require().WithinDuration(
				tC.want.Event.Data.Datetime.AsTime(),
				resp.Event.Data.Datetime.AsTime(),
				time.Second,
				"datetime mismatch",
			)
			s.Require().Equal(
				int(tC.want.Event.Data.Duration.AsDuration().Seconds()),
				int(resp.Event.Data.Duration.AsDuration().Seconds()),
				"duration mismatch",
			)
			s.Require().Equal(
				int(tC.want.Event.Data.RemindIn.AsDuration().Seconds()),
				int(resp.Event.Data.RemindIn.AsDuration().Seconds()),
				"remindIn mismatch",
			)
		})
	}
}

//nolint:funlen
func (s *ServerSuite) TestCreateEvent() {
	input := &types.EventData{
		Title:       "Test Event",
		Datetime:    time.Now(),
		Duration:    time.Hour,
		Description: "Test Description",
		UserID:      basicUserID,
		RemindIn:    time.Minute * 30,
	}
	expectedOutput := &pb.CreateEventResponse{
		Event: &pb.Event{
			Id: uuid.New().String(),
			Data: &pb.EventData{
				Title:       input.Title,
				Datetime:    timestamppb.New(input.Datetime),
				Duration:    durationpb.New(input.Duration),
				Description: input.Description,
				UserId:      input.UserID,
				RemindIn:    durationpb.New(input.RemindIn),
			},
		},
	}
	reqData := expectedOutput.Event.Data

	prepareData := func() *pb.EventData {
		data := &pb.EventData{
			Title:       reqData.Title,
			Datetime:    reqData.Datetime,
			Duration:    reqData.Duration,
			Description: reqData.Description,
			UserId:      reqData.UserId,
			RemindIn:    reqData.RemindIn,
		}
		return data
	}

	testCases := []struct {
		name         string
		req          *pb.CreateEventRequest
		mockApp      func(*mocks.Application)
		want         *pb.CreateEventResponse
		expectedCode codes.Code
	}{
		{
			name: "success",
			req:  &pb.CreateEventRequest{Data: reqData},
			mockApp: func(m *mocks.Application) {
				m.On("CreateEvent", mock.Anything, mock.Anything).Return(
					&types.Event{ID: uuid.New(), EventData: *input},
					nil,
				).Once()
			},
			want:         expectedOutput,
			expectedCode: codes.OK,
		},
		{
			name: "empty fields/title",
			req: func() *pb.CreateEventRequest {
				data := prepareData()
				data.Title = ""
				return &pb.CreateEventRequest{Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("CreateEvent", mock.Anything, mock.Anything).Return(
					nil,
					projectErrors.ErrInvalidFieldData,
				).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "empty fields/duration",
			req: func() *pb.CreateEventRequest {
				data := prepareData()
				data.Duration = nil
				return &pb.CreateEventRequest{Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("CreateEvent", mock.Anything, mock.Anything).Return(
					nil,
					projectErrors.ErrInvalidFieldData,
				).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "empty fields/user_id",
			req: func() *pb.CreateEventRequest {
				data := prepareData()
				data.UserId = ""
				return &pb.CreateEventRequest{Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("CreateEvent", mock.Anything, mock.Anything).Return(
					nil,
					projectErrors.ErrInvalidFieldData,
				).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "empty fields/datetime",
			req: func() *pb.CreateEventRequest {
				data := prepareData()
				data.Datetime = nil
				return &pb.CreateEventRequest{Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("CreateEvent", mock.Anything, mock.Anything).Return(
					nil,
					projectErrors.ErrInvalidFieldData,
				).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "empty fields/description",
			req: func() *pb.CreateEventRequest {
				data := prepareData()
				data.Description = ""
				return &pb.CreateEventRequest{Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("CreateEvent", mock.Anything, mock.Anything).Return(
					&types.Event{ID: uuid.New(), EventData: *input}, nil,
				).Once()
			},
			expectedCode: codes.OK,
			want:         expectedOutput,
		},
		{
			name: "empty fields/remind_in",
			req: func() *pb.CreateEventRequest {
				data := prepareData()
				data.RemindIn = nil
				return &pb.CreateEventRequest{Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("CreateEvent", mock.Anything, mock.Anything).Return(
					&types.Event{ID: uuid.New(), EventData: *input}, nil,
				).Once()
			},
			expectedCode: codes.OK,
			want:         expectedOutput,
		},
		{
			name: "date busy",
			req:  &pb.CreateEventRequest{Data: reqData},
			mockApp: func(m *mocks.Application) {
				m.On("CreateEvent", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrDateBusy).Once()
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "internal error",
			req:  &pb.CreateEventRequest{Data: reqData},
			mockApp: func(m *mocks.Application) {
				m.On("CreateEvent", mock.Anything, mock.Anything).Return(nil, errors.New("unexpected error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.mockApp(s.app)
			s.loggerMocks(s.T())
			resp, err := s.client.CreateEvent(context.Background(), tC.req)

			if tC.expectedCode != codes.OK {
				s.Require().Error(err, "expected error, got nil")
				s.Require().Nil(resp, "expected nil response on error, got non-nil")
				s.Require().Equal(tC.expectedCode, status.Code(err), "unexpected error code")
				return
			}

			s.Require().NoError(err, "unexpected error on GetEvent")
			s.Require().NotNil(resp, "got nil response, expected non-nil")
			s.Require().Equal(tC.want.Event.Data.Title, resp.Event.Data.Title, "title mismatch")
			s.Require().Equal(tC.want.Event.Data.Description, resp.Event.Data.Description, "description mismatch")
			s.Require().Equal(tC.want.Event.Data.UserId, resp.Event.Data.UserId, "user ID mismatch")
			s.Require().WithinDuration(
				tC.want.Event.Data.Datetime.AsTime(),
				resp.Event.Data.Datetime.AsTime(),
				time.Second,
				"datetime mismatch",
			)
			s.Require().Equal(
				int(tC.want.Event.Data.Duration.AsDuration().Seconds()),
				int(resp.Event.Data.Duration.AsDuration().Seconds()),
				"duration mismatch",
			)
			s.Require().Equal(
				int(tC.want.Event.Data.RemindIn.AsDuration().Seconds()),
				int(resp.Event.Data.RemindIn.AsDuration().Seconds()),
				"remindIn mismatch",
			)
		})
	}
}

//nolint:funlen
func (s *ServerSuite) TestUpdateEvent() {
	id := uuid.New()
	input := &types.EventData{
		Title:       "Test Event",
		Datetime:    time.Now(),
		Duration:    time.Hour,
		Description: "Test Description",
		UserID:      basicUserID,
		RemindIn:    time.Minute * 30,
	}
	expectedOutput := &pb.UpdateEventResponse{
		Event: &pb.Event{
			Id: id.String(),
			Data: &pb.EventData{
				Title:       input.Title,
				Datetime:    timestamppb.New(input.Datetime),
				Duration:    durationpb.New(input.Duration),
				Description: input.Description,
				UserId:      input.UserID,
				RemindIn:    durationpb.New(input.RemindIn),
			},
		},
	}
	reqData := expectedOutput.Event.Data

	prepareData := func() *pb.EventData {
		data := &pb.EventData{
			Title:       reqData.Title,
			Datetime:    reqData.Datetime,
			Duration:    reqData.Duration,
			Description: reqData.Description,
			UserId:      reqData.UserId,
			RemindIn:    reqData.RemindIn,
		}
		return data
	}

	testCases := []struct {
		name         string
		req          *pb.UpdateEventRequest
		mockApp      func(*mocks.Application)
		want         *pb.UpdateEventResponse
		expectedCode codes.Code
	}{
		{
			name: "success",
			req:  &pb.UpdateEventRequest{Id: id.String(), Data: reqData},
			mockApp: func(m *mocks.Application) {
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(
					&types.Event{ID: id, EventData: *input},
					nil,
				).Once()
			},
			want:         expectedOutput,
			expectedCode: codes.OK,
		},
		{
			name: "empty fields/id",
			req: func() *pb.UpdateEventRequest {
				data := prepareData()
				return &pb.UpdateEventRequest{Id: "", Data: data}
			}(),
			mockApp: func(_ *mocks.Application) {
				// No calls expected here.
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "empty fields/title",
			req: func() *pb.UpdateEventRequest {
				data := prepareData()
				data.Title = ""
				return &pb.UpdateEventRequest{Id: id.String(), Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(
					nil,
					projectErrors.ErrInvalidFieldData,
				).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "empty fields/duration",
			req: func() *pb.UpdateEventRequest {
				data := prepareData()
				data.Duration = nil
				return &pb.UpdateEventRequest{Id: id.String(), Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(
					nil,
					projectErrors.ErrInvalidFieldData,
				).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "empty fields/user_id",
			req: func() *pb.UpdateEventRequest {
				data := prepareData()
				data.UserId = ""
				return &pb.UpdateEventRequest{Id: id.String(), Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(
					nil,
					projectErrors.ErrInvalidFieldData,
				).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "empty fields/datetime",
			req: func() *pb.UpdateEventRequest {
				data := prepareData()
				data.Datetime = nil
				return &pb.UpdateEventRequest{Id: id.String(), Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(
					nil,
					projectErrors.ErrInvalidFieldData,
				).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "empty fields/description",
			req: func() *pb.UpdateEventRequest {
				data := prepareData()
				data.Description = ""
				return &pb.UpdateEventRequest{Id: id.String(), Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(
					&types.Event{ID: id, EventData: *input}, nil,
				).Once()
			},
			expectedCode: codes.OK,
			want:         expectedOutput,
		},
		{
			name: "empty fields/remind_in",
			req: func() *pb.UpdateEventRequest {
				data := prepareData()
				data.RemindIn = nil
				return &pb.UpdateEventRequest{Id: id.String(), Data: data}
			}(),
			mockApp: func(m *mocks.Application) {
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(
					&types.Event{ID: id, EventData: *input}, nil,
				).Once()
			},
			expectedCode: codes.OK,
			want:         expectedOutput,
		},
		{
			name: "date busy",
			req:  &pb.UpdateEventRequest{Id: id.String(), Data: reqData},
			mockApp: func(m *mocks.Application) {
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrDateBusy).Once()
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "permission denied",
			req:  &pb.UpdateEventRequest{Id: id.String(), Data: reqData},
			mockApp: func(m *mocks.Application) {
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrPermissionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "internal error",
			req:  &pb.UpdateEventRequest{Id: id.String(), Data: reqData},
			mockApp: func(m *mocks.Application) {
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(nil, errors.New("unexpected error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.mockApp(s.app)
			s.loggerMocks(s.T())
			resp, err := s.client.UpdateEvent(context.Background(), tC.req)

			if tC.expectedCode != codes.OK {
				s.Require().Error(err, "expected error, got nil")
				s.Require().Nil(resp, "expected nil response on error, got non-nil")
				s.Require().Equal(tC.expectedCode, status.Code(err), "unexpected error code")
				return
			}

			s.Require().NoError(err, "unexpected error on GetEvent")
			s.Require().NotNil(resp, "got nil response, expected non-nil")
			s.Require().Equal(tC.want.Event.Data.Title, resp.Event.Data.Title, "title mismatch")
			s.Require().Equal(tC.want.Event.Data.Description, resp.Event.Data.Description, "description mismatch")
			s.Require().Equal(tC.want.Event.Data.UserId, resp.Event.Data.UserId, "user ID mismatch")
			s.Require().WithinDuration(
				tC.want.Event.Data.Datetime.AsTime(),
				resp.Event.Data.Datetime.AsTime(),
				time.Second,
				"datetime mismatch",
			)
			s.Require().Equal(
				int(tC.want.Event.Data.Duration.AsDuration().Seconds()),
				int(resp.Event.Data.Duration.AsDuration().Seconds()),
				"duration mismatch",
			)
			s.Require().Equal(
				int(tC.want.Event.Data.RemindIn.AsDuration().Seconds()),
				int(resp.Event.Data.RemindIn.AsDuration().Seconds()),
				"remindIn mismatch",
			)
		})
	}
}

func (s *ServerSuite) TestDeleteEvent() {
	id := uuid.New()

	testCases := []struct {
		name         string
		req          *pb.DeleteEventRequest
		mockApp      func(*mocks.Application)
		expectedCode codes.Code
	}{
		{
			name: "success",
			req:  &pb.DeleteEventRequest{Id: id.String()},
			mockApp: func(m *mocks.Application) {
				m.On("DeleteEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name: "invalid uuid",
			req:  &pb.DeleteEventRequest{Id: "invalid-uuid"},
			mockApp: func(_ *mocks.Application) {
				// No mock setup as parseUUID will fail before calling Application.
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "event not found",
			req:  &pb.DeleteEventRequest{Id: uuid.New().String()},
			mockApp: func(m *mocks.Application) {
				m.On("DeleteEvent", mock.Anything, mock.Anything).Return(projectErrors.ErrEventNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  &pb.DeleteEventRequest{Id: uuid.New().String()},
			mockApp: func(m *mocks.Application) {
				m.On("DeleteEvent", mock.Anything, mock.Anything).Return(projectErrors.ErrInconsistentState).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.mockApp(s.app)
			s.loggerMocks(s.T())
			resp, err := s.client.DeleteEvent(context.Background(), tC.req)

			if tC.expectedCode != codes.OK {
				s.Require().Error(err, "expected error, got nil")
				s.Require().Nil(resp, "expected nil response on error, got non-nil")
				s.Require().Equal(tC.expectedCode, status.Code(err), "unexpected error code")
				return
			}

			s.Require().NoError(err, "unexpected error on DeleteEvent")
			s.Require().NotNil(resp, "got nil response, expected non-nil")
		})
	}
}

//nolint:funlen
func (s *ServerSuite) TestGetAllUserEvents() {
	userID := basicUserID
	now := time.Now()
	eventID1 := uuid.New()
	eventID2 := uuid.New()

	eventsFromApp := []*types.Event{
		{
			ID: eventID1,
			EventData: types.EventData{
				Title:       "Event 1",
				Datetime:    now,
				Duration:    time.Hour,
				Description: "Description 1",
				UserID:      userID,
				RemindIn:    time.Minute * 15,
			},
		},
		{
			ID: eventID2,
			EventData: types.EventData{
				Title:       "Event 2",
				Datetime:    now.Add(time.Hour),
				Duration:    time.Hour * 2,
				Description: "Description 2",
				UserID:      userID,
				RemindIn:    time.Minute * 30,
			},
		},
	}

	expectedPBEvents := []*pb.Event{
		{
			Id: eventID1.String(),
			Data: &pb.EventData{
				Title:       "Event 1",
				Datetime:    timestamppb.New(now),
				Duration:    durationpb.New(time.Hour),
				Description: "Description 1",
				UserId:      userID,
				RemindIn:    durationpb.New(time.Minute * 15),
			},
		},
		{
			Id: eventID2.String(),
			Data: &pb.EventData{
				Title:       "Event 2",
				Datetime:    timestamppb.New(now.Add(time.Hour)),
				Duration:    durationpb.New(time.Hour * 2),
				Description: "Description 2",
				UserId:      userID,
				RemindIn:    durationpb.New(time.Minute * 30),
			},
		},
	}

	testCases := []struct {
		name         string
		req          *pb.GetAllUserEventsRequest
		mockApp      func(*mocks.Application)
		want         *pb.GetAllUserEventsResponse
		expectedCode codes.Code
	}{
		{
			name: "success",
			req:  &pb.GetAllUserEventsRequest{UserId: userID},
			mockApp: func(m *mocks.Application) {
				m.On("GetAllUserEvents", mock.Anything, mock.Anything).Return(eventsFromApp, nil).Once()
			},
			want: &pb.GetAllUserEventsResponse{
				Events: expectedPBEvents,
			},
			expectedCode: codes.OK,
		},
		{
			name: "empty user_id",
			req:  &pb.GetAllUserEventsRequest{UserId: ""},
			mockApp: func(m *mocks.Application) {
				m.On("GetAllUserEvents", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrInvalidFieldData).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "no events found",
			req:  &pb.GetAllUserEventsRequest{UserId: userID},
			mockApp: func(m *mocks.Application) {
				m.On("GetAllUserEvents", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrEventNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  &pb.GetAllUserEventsRequest{UserId: userID},
			mockApp: func(m *mocks.Application) {
				m.On("GetAllUserEvents", mock.Anything, mock.Anything).Return(nil, errors.New("unexpected error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tC := range testCases {
		s.Run(tC.name, func() {
			tC.mockApp(s.app)
			s.loggerMocks(s.T())

			resp, err := s.client.GetAllUserEvents(context.Background(), tC.req)

			if tC.expectedCode != codes.OK {
				s.Require().Error(err, "expected error, got nil")
				s.Require().Nil(resp, "expected nil response on error, got non-nil")
				s.Require().Equal(tC.expectedCode, status.Code(err), "unexpected error code")
				return
			}

			s.Require().NoError(err, "unexpected error on GetAllUserEvents")
			s.Require().NotNil(resp, "got nil response, expected non-nil")

			s.Require().Len(resp.Events, len(tC.want.Events), "unexpected number of events")

			for i := range resp.Events {
				s.Require().Equal(tC.want.Events[i].Id, resp.Events[i].Id, "event ID mismatch at index %d", i)
				s.Require().Equal(
					tC.want.Events[i].Data.Title,
					resp.Events[i].Data.Title,
					"title mismatch at index %d",
					i,
				)
				s.Require().Equal(
					tC.want.Events[i].Data.Description,
					resp.Events[i].Data.Description,
					"description mismatch at index %d",
					i,
				)
				s.Require().Equal(
					tC.want.Events[i].Data.UserId,
					resp.Events[i].Data.UserId,
					"user ID mismatch at index %d",
					i,
				)
				s.Require().WithinDuration(
					tC.want.Events[i].Data.Datetime.AsTime(),
					resp.Events[i].Data.Datetime.AsTime(),
					time.Second,
					"datetime mismatch at index %d",
					i,
				)
				s.Require().Equal(
					int(tC.want.Events[i].Data.Duration.AsDuration().Seconds()),
					int(resp.Events[i].Data.Duration.AsDuration().Seconds()),
					"duration mismatch at index %d",
					i,
				)
				s.Require().Equal(
					int(tC.want.Events[i].Data.RemindIn.AsDuration().Seconds()),
					int(resp.Events[i].Data.RemindIn.AsDuration().Seconds()),
					"remind_in mismatch at index %d",
					i,
				)
			}
		})
	}
}

func (s *ServerSuite) TestGetEventsForPeriod() {
	userID := basicUserID
	now := time.Now()

	eventsFromApp := []*types.Event{
		{
			ID: uuid.New(),
			EventData: types.EventData{
				Title:       "Event 1",
				Datetime:    now,
				Duration:    time.Hour,
				Description: "Description 1",
				UserID:      userID,
				RemindIn:    time.Minute * 15,
			},
		},
		{
			ID: uuid.New(),
			EventData: types.EventData{
				Title:       "Event 2",
				Datetime:    now.Add(time.Hour),
				Duration:    time.Hour * 2,
				Description: "Description 2",
				UserID:      userID,
				RemindIn:    time.Minute * 30,
			},
		},
	}

	pbStartDate := timestamppb.New(now)
	pbEndDate := timestamppb.New(now.Add(24 * time.Hour))

	req := &pb.GetEventsForPeriodRequest{
		StartDate: pbStartDate,
		EndDate:   pbEndDate,
		UserId:    &userID,
	}

	testCases := []struct {
		name         string
		req          *pb.GetEventsForPeriodRequest
		mockApp      func(*mocks.Application)
		expectedCode codes.Code
		expectedLen  int
	}{
		{
			name: "success",
			req:  req,
			mockApp: func(m *mocks.Application) {
				m.On("GetEventsForPeriod", mock.Anything, mock.Anything).Return(eventsFromApp, nil).Once()
			},
			expectedCode: codes.OK,
			expectedLen:  len(eventsFromApp),
		},
		{
			name: "empty user_id",
			req: &pb.GetEventsForPeriodRequest{
				StartDate: pbStartDate,
				EndDate:   pbEndDate,
				UserId:    nil,
			},
			mockApp: func(m *mocks.Application) {
				m.On("GetEventsForPeriod", mock.Anything, mock.Anything).Return(eventsFromApp, nil).Once()
			},
			expectedCode: codes.OK,
			expectedLen:  len(eventsFromApp),
		},
		{
			name: "invalid start date",
			req: &pb.GetEventsForPeriodRequest{
				StartDate: nil,
				EndDate:   pbEndDate,
				UserId:    &userID,
			},
			mockApp: func(m *mocks.Application) {
				m.On("GetEventsForPeriod", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrInvalidFieldData).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "invalid end date",
			req: &pb.GetEventsForPeriodRequest{
				StartDate: pbStartDate,
				EndDate:   nil,
				UserId:    &userID,
			},
			mockApp: func(m *mocks.Application) {
				m.On("GetEventsForPeriod", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrInvalidFieldData).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "no events found",
			req:  req,
			mockApp: func(m *mocks.Application) {
				m.On("GetEventsForPeriod", mock.Anything, mock.Anything).Return([]*types.Event{}, nil).Once()
			},
			expectedCode: codes.OK,
			expectedLen:  0,
		},
		{
			name: "internal error",
			req:  req,
			mockApp: func(m *mocks.Application) {
				m.On("GetEventsForPeriod", mock.Anything, mock.Anything).Return(nil, errors.New("unexpected error")).Once()
			},
			expectedCode: codes.Internal,
			expectedLen:  0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mockApp(s.app)
			s.loggerMocks(s.T())

			resp, err := s.client.GetEventsForPeriod(context.Background(), tc.req)

			if tc.expectedCode != codes.OK {
				s.Require().Error(err)
				s.Require().Nil(resp)
				s.Require().Equal(tc.expectedCode, status.Code(err))
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Len(resp.Events, tc.expectedLen)
		})
	}
}

func (s *ServerSuite) TestGetEventsForDay_Week_Month() {
	userID := basicUserID
	now := time.Now()

	eventsFromApp := []*types.Event{
		{
			ID: uuid.New(),
			EventData: types.EventData{
				Title:       "Event 1",
				Datetime:    now,
				Duration:    time.Hour,
				Description: "Description 1",
				UserID:      userID,
				RemindIn:    time.Minute * 15,
			},
		},
		{
			ID: uuid.New(),
			EventData: types.EventData{
				Title:       "Event 2",
				Datetime:    now.Add(time.Hour),
				Duration:    time.Hour * 2,
				Description: "Description 2",
				UserID:      userID,
				RemindIn:    time.Minute * 30,
			},
		},
	}

	pbDate := timestamppb.New(now)

	req := &pb.GetEventsForDayRequest{
		Date:   pbDate,
		UserId: &userID,
	}

	testCases := []struct {
		name         string
		req          *pb.GetEventsForDayRequest
		mockApp      func(*mocks.Application)
		expectedCode codes.Code
		expectedLen  int
	}{
		{
			name: "success",
			req:  req,
			mockApp: func(m *mocks.Application) {
				m.On("ListEvents", mock.Anything, mock.Anything).Return(eventsFromApp, nil).Once()
			},
			expectedCode: codes.OK,
			expectedLen:  len(eventsFromApp),
		},
		{
			name: "empty user_id",
			req: &pb.GetEventsForDayRequest{
				Date:   pbDate,
				UserId: nil,
			},
			mockApp: func(m *mocks.Application) {
				m.On("ListEvents", mock.Anything, mock.Anything).Return(eventsFromApp, nil).Once()
			},
			expectedCode: codes.OK,
			expectedLen:  len(eventsFromApp),
		},
		{
			name: "invalid date",
			req: &pb.GetEventsForDayRequest{
				Date:   nil,
				UserId: &userID,
			},
			mockApp: func(m *mocks.Application) {
				m.On("ListEvents", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrInvalidFieldData).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "no events found",
			req:  req,
			mockApp: func(m *mocks.Application) {
				m.On("ListEvents", mock.Anything, mock.Anything).Return(nil, projectErrors.ErrEventNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  req,
			mockApp: func(m *mocks.Application) {
				m.On("ListEvents", mock.Anything, mock.Anything).Return(nil, errors.New("some error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mockApp(s.app)
			s.loggerMocks(s.T())

			resp, err := s.client.GetEventsForDay(context.Background(), tc.req)

			if tc.expectedCode != codes.OK {
				s.Require().Error(err)
				s.Require().Nil(resp)
				s.Require().Equal(tc.expectedCode, status.Code(err))
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Len(resp.Events, tc.expectedLen)
		})
	}
}
