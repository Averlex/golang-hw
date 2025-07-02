package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1"                         //nolint:depguard,nolintlint
	calendarGRPC "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/server/calendar/grpc" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/server/calendar/grpc/mocks"        //nolint:depguard,nolintlint
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors"                   //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"                                  //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                                            //nolint:depguard,nolintlint
	"github.com/stretchr/testify/mock"                                                                  //nolint:depguard,nolintlint
	"github.com/stretchr/testify/require"                                                               //nolint:depguard,nolintlint
	"github.com/stretchr/testify/suite"                                                                 //nolint:depguard,nolintlint
	"google.golang.org/grpc"                                                                            //nolint:depguard,nolintlint
	"google.golang.org/grpc/codes"                                                                      //nolint:depguard,nolintlint
	"google.golang.org/grpc/credentials/insecure"                                                       //nolint:depguard,nolintlint
	"google.golang.org/grpc/status"                                                                     //nolint:depguard,nolintlint
	"google.golang.org/grpc/test/bufconn"                                                               //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/durationpb"                                                 //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/timestamppb"                                                //nolint:depguard,nolintlint
)

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
	s.loggerMocks(s.T())

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
			UserID:      "user1",
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
