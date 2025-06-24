//nolint:lll,nolintlint

package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/calendar/dto"   //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"          //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                    //nolint:depguard,nolintlint
	"google.golang.org/grpc"                                                    //nolint:depguard,nolintlint
	"google.golang.org/grpc/metadata"                                           //nolint:depguard,nolintlint
)

// CreateEvent validates the request data and tries to create a new event in the storage.
func (s *Server) CreateEvent(ctx context.Context, event *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	var obj dto.CreateEventInput

	// Request data preprocessing.
	if event != nil && event.Data != nil {
		var reqDuration time.Duration
		if setDuration(event.Data.Duration) != nil {
			reqDuration = *setDuration(event.Data.Duration)
		}

		obj = dto.CreateEventInput{
			Title:       event.Data.Title,
			Datetime:    setTime(event.Data.Datetime),
			Duration:    reqDuration,
			Description: setDesctription(event.Data.Description),
			RemindIn:    setDuration(event.Data.RemindIn),
			UserID:      event.Data.UserId,
		}
	}

	res, err := s.a.CreateEvent(ctx, &obj)
	st := s.wrapError(ctx, err)

	_ = grpc.SetHeader(ctx, metadata.MD{}) // To ensure that the error is sent to the client.

	if err != nil {
		return nil, st.Err()
	}

	return &pb.CreateEventResponse{
		Event: fromInternalEvent(res),
	}, nil
}

// UpdateEvent validates the request data and tries to update an existing event in the storage.
func (s *Server) UpdateEvent(ctx context.Context, data *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	var obj dto.UpdateEventInput
	var err error
	var res *types.Event

	// Request data preprocessing.
	if data != nil && data.Data != nil {
		var reqDuration time.Duration
		if setDuration(data.Data.Duration) != nil {
			reqDuration = *setDuration(data.Data.Duration)
		}
		id, idErr := uuid.Parse(data.Id)
		if idErr != nil {
			id = uuid.Nil
			err = fmt.Errorf("invalid id format in request data: %w", idErr)
		}
		datetime := setTime(data.Data.Datetime)

		obj = dto.UpdateEventInput{
			ID:          id,
			Title:       &data.Data.Title,
			Datetime:    &datetime,
			Duration:    &reqDuration,
			Description: setDesctription(data.Data.Description),
			RemindIn:    setDuration(data.Data.RemindIn),
			UserID:      &data.Data.UserId,
		}
	}

	if err != nil {
		res, err = s.a.UpdateEvent(ctx, &obj)
	}
	st := s.wrapError(ctx, err)

	_ = grpc.SetHeader(ctx, metadata.MD{}) // To ensure that the error is sent to the client.

	if err != nil {
		return nil, st.Err()
	}

	return &pb.UpdateEventResponse{
		Event: fromInternalEvent(res),
	}, nil
}

// func (s *Server) DeleteEvent(context.Context, *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method DeleteEvent not implemented")
// }

// func (s *Server) GetEvent(context.Context, *pb.GetEventRequest) (*pb.GetEventResponse, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method GetEvent not implemented")
// }

// func (s *Server) GetAllUserEvents(context.Context, *pb.GetAllUserEventsRequest) (
// *pb.GetAllUserEventsResponse,
// error,
// ) {
// 	return nil, status.Errorf(codes.Unimplemented, "method GetAllUserEvents not implemented")
// }

// func (s *Server) GetEventsForDay(context.Context, *pb.GetEventsForDayRequest) (*pb.GetEventsForDayResponse, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method GetEventsForDay not implemented")
// }

// func (s *Server) GetEventsForWeek(context.Context, *pb.GetEventsForWeekRequest) (
// *pb.GetEventsForWeekResponse,
// error,
// ) {
// 	return nil, status.Errorf(codes.Unimplemented, "method GetEventsForWeek not implemented")
// }

// func (s *Server) GetEventsForMonth(context.Context, *pb.GetEventsForMonthRequest) (
// *pb.GetEventsForMonthResponse,
// error,
// ) {
// 	return nil, status.Errorf(codes.Unimplemented, "method GetEventsForMonth not implemented")
// }

// func (s *Server) GetEventsForPeriod(context.Context, *pb.GetEventsForPeriodRequest) (
// *pb.GetEventsForPeriodResponse,
// error,
// ) {
// 	return nil, status.Errorf(codes.Unimplemented, "method GetEventsForPeriod not implemented")
// }
