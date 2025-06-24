//nolint:lll,nolintlint

package grpc

import (
	"context"
	"time"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/calendar/dto"   //nolint:depguard,nolintlint
)

// CreateEvent validates the request data and tries to create a new event in the storage.
func (s *Server) CreateEvent(ctx context.Context, event *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	var obj dto.CreateEventInput

	if event != nil {
		if event.Data != nil {
			// Request data preprocessing.
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
	}

	res, err := s.a.CreateEvent(ctx, &obj)
	if err != nil {
		return &pb.CreateEventResponse{
			Event:  nil,
			Status: s.wrapError(ctx, err),
		}, nil
	}

	return &pb.CreateEventResponse{
		Event:  fromInternalEvent(res),
		Status: s.wrapError(ctx, err),
	}, nil
}

// func (s *Server) UpdateEvent(context.Context, *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method UpdateEvent not implemented")
// }

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
