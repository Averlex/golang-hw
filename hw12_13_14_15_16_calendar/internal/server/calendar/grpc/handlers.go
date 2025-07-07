//nolint:lll,nolintlint

package grpc

import (
	"context"
	"time"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/calendar/dto"   //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/timestamppb"                        //nolint:depguard,nolintlint
)

// CreateEvent validates the request data and tries to create a new event in the storage.
func (s *Server) CreateEvent(ctx context.Context, event *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	var duration time.Duration
	reqDuration := setDuration(event.Data.Duration)
	if reqDuration != nil {
		duration = *reqDuration
	}
	obj := dto.CreateEventInput{
		Title:       event.Data.Title,
		Datetime:    setTime(event.Data.Datetime),
		Duration:    duration,
		Description: setDesctription(event.Data.Description),
		RemindIn:    setDuration(event.Data.RemindIn),
		UserID:      event.Data.UserId,
	}

	res, err := s.a.CreateEvent(ctx, &obj)
	if err != nil {
		return nil, s.handleError(ctx, err).Err()
	}

	return &pb.CreateEventResponse{
		Event: fromInternalEvent(res),
	}, nil
}

// UpdateEvent validates the request data and tries to update an existing event in the storage.
func (s *Server) UpdateEvent(ctx context.Context, data *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	id, err := parseUUID(data.Id)
	if err != nil {
		return nil, s.handleError(ctx, err).Err()
	}

	datetime := setTime(data.Data.Datetime)
	reqDuration := setDuration(data.Data.Duration)
	obj := dto.UpdateEventInput{
		ID:          id,
		Title:       &data.Data.Title,
		Datetime:    &datetime,
		Duration:    reqDuration,
		Description: setDesctription(data.Data.Description),
		RemindIn:    setDuration(data.Data.RemindIn),
		UserID:      &data.Data.UserId,
	}

	res, err := s.a.UpdateEvent(ctx, &obj)
	if err != nil {
		return nil, s.handleError(ctx, err).Err()
	}

	return &pb.UpdateEventResponse{
		Event: fromInternalEvent(res),
	}, nil
}

// DeleteEvent tries to delete the Event with the given ID from the storage.
func (s *Server) DeleteEvent(ctx context.Context, data *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	id, err := parseUUID(data.Id)
	if err != nil {
		return nil, s.handleError(ctx, err).Err()
	}

	err = s.a.DeleteEvent(ctx, id.String())
	if err != nil {
		return nil, s.handleError(ctx, err).Err()
	}

	return &pb.DeleteEventResponse{}, nil
}

// GetEvent tries to get the Event with the given ID from the storage.
func (s *Server) GetEvent(ctx context.Context, data *pb.GetEventRequest) (*pb.GetEventResponse, error) {
	id, err := parseUUID(data.Id)
	if err != nil {
		return nil, s.handleError(ctx, err).Err()
	}

	res, err := s.a.GetEvent(ctx, id.String())
	if err != nil {
		return nil, s.handleError(ctx, err).Err()
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
	res, err := s.a.GetAllUserEvents(ctx, data.UserId)
	if err != nil {
		return nil, s.handleError(ctx, err).Err()
	}

	return &pb.GetAllUserEventsResponse{
		Events: convertEventsToPB(res),
	}, nil
}

// getEventsByPeriod is a helper function to get events for a specific period.
func (s *Server) getEventsByPeriod(
	ctx context.Context,
	date *timestamppb.Timestamp,
	userID *string,
	period dto.Period,
) ([]*pb.Event, error) {
	obj := dto.DateFilterInput{
		Date:   setTime(date),
		UserID: userID,
		Period: period,
	}

	res, err := s.a.ListEvents(ctx, &obj)
	if err != nil {
		return nil, s.handleError(ctx, err).Err()
	}

	return convertEventsToPB(res), nil
}

// GetEventsForDay is trying to get all events for a given day from the storage.
func (s *Server) GetEventsForDay(
	ctx context.Context,
	data *pb.GetEventsForDayRequest,
) (*pb.GetEventsForDayResponse, error) {
	events, err := s.getEventsByPeriod(ctx, data.Date, data.UserId, dto.Day)
	if err != nil {
		return nil, err
	}

	return &pb.GetEventsForDayResponse{
		Events: events,
	}, nil
}

// GetEventsForWeek is trying to get all events for a given week from the storage.
func (s *Server) GetEventsForWeek(
	ctx context.Context,
	data *pb.GetEventsForWeekRequest,
) (*pb.GetEventsForWeekResponse, error) {
	events, err := s.getEventsByPeriod(ctx, data.Date, data.UserId, dto.Week)
	if err != nil {
		return nil, err
	}

	return &pb.GetEventsForWeekResponse{
		Events: events,
	}, nil
}

// GetEventsForMonth is trying to get all events for a given month from the storage.
func (s *Server) GetEventsForMonth(
	ctx context.Context,
	data *pb.GetEventsForMonthRequest,
) (*pb.GetEventsForMonthResponse, error) {
	events, err := s.getEventsByPeriod(ctx, data.Date, data.UserId, dto.Month)
	if err != nil {
		return nil, err
	}

	return &pb.GetEventsForMonthResponse{
		Events: events,
	}, nil
}

// GetEventsForPeriod is trying to get all events for a given period from the storage.
func (s *Server) GetEventsForPeriod(
	ctx context.Context,
	data *pb.GetEventsForPeriodRequest,
) (*pb.GetEventsForPeriodResponse, error) {
	obj := dto.DateRangeInput{
		DateStart: setTime(data.StartDate),
		DateEnd:   setTime(data.EndDate),
		UserID:    data.UserId,
	}

	res, err := s.a.GetEventsForPeriod(ctx, &obj)
	if err != nil {
		return nil, s.handleError(ctx, err).Err()
	}

	return &pb.GetEventsForPeriodResponse{
		Events: convertEventsToPB(res),
	}, nil
}
