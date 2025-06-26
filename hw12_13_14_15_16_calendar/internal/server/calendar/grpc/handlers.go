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

	// Request data preprocessing.
	if data != nil {
		userID = data.UserId
	}

	res, err = s.a.GetAllUserEvents(ctx, userID)

	st := s.wrapError(ctx, err)

	_ = grpc.SetHeader(ctx, metadata.MD{}) // To ensure that the error is sent to the client.

	if err != nil {
		return nil, st.Err()
	}

	pbRes := make([]*pb.Event, len(res))
	for i := range res {
		pbRes[i] = fromInternalEvent(res[i])
	}

	return &pb.GetAllUserEventsResponse{
		Events: pbRes,
	}, nil
}

// GetEventsForDay is trying to get all events for a given day from the storage.
//
//nolint:dupl
func (s *Server) GetEventsForDay(
	ctx context.Context,
	data *pb.GetEventsForDayRequest,
) (*pb.GetEventsForDayResponse, error) {
	var obj dto.DateFilterInput
	var res []*types.Event
	var err error

	// Request data preprocessing.
	if data != nil {
		obj.Date = setTime(data.Date)
		obj.UserID = data.UserId
		obj.Period = dto.Day
	}

	res, err = s.a.ListEvents(ctx, &obj)

	st := s.wrapError(ctx, err)

	_ = grpc.SetHeader(ctx, metadata.MD{}) // To ensure that the error is sent to the client.

	if err != nil {
		return nil, st.Err()
	}

	pbRes := make([]*pb.Event, len(res))
	for i := range res {
		pbRes[i] = fromInternalEvent(res[i])
	}

	return &pb.GetEventsForDayResponse{
		Events: pbRes,
	}, nil
}

// GetEventsForWeek is trying to get all events for a given week from the storage.
//
//nolint:dupl
func (s *Server) GetEventsForWeek(
	ctx context.Context,
	data *pb.GetEventsForWeekRequest,
) (*pb.GetEventsForWeekResponse, error) {
	var obj dto.DateFilterInput
	var res []*types.Event
	var err error

	// Request data preprocessing.
	if data != nil {
		obj.Date = setTime(data.Date)
		obj.UserID = data.UserId
		obj.Period = dto.Week
	}

	res, err = s.a.ListEvents(ctx, &obj)

	st := s.wrapError(ctx, err)

	_ = grpc.SetHeader(ctx, metadata.MD{}) // To ensure that the error is sent to the client.

	if err != nil {
		return nil, st.Err()
	}

	pbRes := make([]*pb.Event, len(res))
	for i := range res {
		pbRes[i] = fromInternalEvent(res[i])
	}

	return &pb.GetEventsForWeekResponse{
		Events: pbRes,
	}, nil
}

// GetEventsForMonth is trying to get all events for a given month from the storage.
//
//nolint:dupl
func (s *Server) GetEventsForMonth(
	ctx context.Context,
	data *pb.GetEventsForMonthRequest,
) (*pb.GetEventsForMonthResponse, error) {
	var obj dto.DateFilterInput
	var res []*types.Event
	var err error

	// Request data preprocessing.
	if data != nil {
		obj.Date = setTime(data.Date)
		obj.UserID = data.UserId
		obj.Period = dto.Month
	}

	res, err = s.a.ListEvents(ctx, &obj)

	st := s.wrapError(ctx, err)

	_ = grpc.SetHeader(ctx, metadata.MD{}) // To ensure that the error is sent to the client.

	if err != nil {
		return nil, st.Err()
	}

	pbRes := make([]*pb.Event, len(res))
	for i := range res {
		pbRes[i] = fromInternalEvent(res[i])
	}

	return &pb.GetEventsForMonthResponse{
		Events: pbRes,
	}, nil
}

// GetEventsForPeriod is trying to get all events for a given period from the storage.
func (s *Server) GetEventsForPeriod(
	ctx context.Context,
	data *pb.GetEventsForPeriodRequest,
) (*pb.GetEventsForPeriodResponse, error) {
	var obj dto.DateRangeInput
	var res []*types.Event
	var err error

	// Request data preprocessing.
	if data != nil {
		obj.DateStart = setTime(data.StartDate)
		obj.DateEnd = setTime(data.EndDate)
		obj.UserID = data.UserId
	}

	res, err = s.a.GetEventsForPeriod(ctx, &obj)

	st := s.wrapError(ctx, err)

	_ = grpc.SetHeader(ctx, metadata.MD{}) // To ensure that the error is sent to the client.

	if err != nil {
		return nil, st.Err()
	}

	pbRes := make([]*pb.Event, len(res))
	for i := range res {
		pbRes[i] = fromInternalEvent(res[i])
	}

	return &pb.GetEventsForPeriodResponse{
		Events: pbRes,
	}, nil
}
