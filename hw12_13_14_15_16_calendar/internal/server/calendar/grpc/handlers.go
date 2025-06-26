//nolint:lll,nolintlint

package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1"       //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/calendar/dto"         //nolint:depguard,nolintlint
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                          //nolint:depguard,nolintlint
	"google.golang.org/grpc"                                                          //nolint:depguard,nolintlint
	"google.golang.org/grpc/metadata"                                                 //nolint:depguard,nolintlint
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
			err = fmt.Errorf("%w: invalid id format in request data: %w", projectErrors.ErrInvalidFieldData, idErr)
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

	if err == nil {
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

// DeleteEvent tries to delete the Event with the given ID from the storage.
func (s *Server) DeleteEvent(ctx context.Context, data *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	var id uuid.UUID
	var err error

	// Request data preprocessing.
	if data != nil {
		var idErr error
		id, idErr = uuid.Parse(data.Id)
		if idErr != nil {
			id = uuid.Nil
			err = fmt.Errorf("%w: invalid id format in request data: %w", projectErrors.ErrInvalidFieldData, idErr)
		}
	}

	if err == nil {
		err = s.a.DeleteEvent(ctx, id.String())
	}
	st := s.wrapError(ctx, err)

	_ = grpc.SetHeader(ctx, metadata.MD{}) // To ensure that the error is sent to the client.

	if err != nil {
		return nil, st.Err()
	}

	return &pb.DeleteEventResponse{}, nil
}

// GetEvent tries to get the Event with the given ID from the storage.
func (s *Server) GetEvent(ctx context.Context, data *pb.GetEventRequest) (*pb.GetEventResponse, error) {
	var id uuid.UUID
	var res *types.Event
	var err error

	// Request data preprocessing.
	if data != nil {
		var idErr error
		id, idErr = uuid.Parse(data.Id)
		if idErr != nil {
			id = uuid.Nil
			err = fmt.Errorf("%w: invalid id format in request data: %w", projectErrors.ErrInvalidFieldData, idErr)
		}
	}

	if err == nil {
		res, err = s.a.GetEvent(ctx, id.String())
	}
	st := s.wrapError(ctx, err)

	_ = grpc.SetHeader(ctx, metadata.MD{}) // To ensure that the error is sent to the client.

	if err != nil {
		return nil, st.Err()
	}

	return &pb.GetEventResponse{
		Event: fromInternalEvent(res),
	}, nil
}

// GetAllUserEvents is trying to get all events for a given user ID from the storage.
func (s *Server) GetAllUserEvents(ctx context.Context, data *pb.GetAllUserEventsRequest) (
	*pb.GetAllUserEventsResponse,
	error,
) {
	var userID string
	var res []*types.Event
	var err error
	pbRes := make([]*pb.Event, 0)

	// Request data preprocessing.
	if data != nil {
		userID = data.UserId
	}

	if err == nil {
		res, err = s.a.GetAllUserEvents(ctx, userID)
	}
	st := s.wrapError(ctx, err)

	_ = grpc.SetHeader(ctx, metadata.MD{}) // To ensure that the error is sent to the client.

	if err != nil {
		return nil, st.Err()
	}

	for i := range res {
		pbRes = append(pbRes, fromInternalEvent(res[i]))
	}

	return &pb.GetAllUserEventsResponse{
		Events: pbRes,
	}, nil
}

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
