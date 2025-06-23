package grpc

import (
	"fmt"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1"       //nolint:depguard,nolintlint
	packageErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                          //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/durationpb"                               //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/timestamppb"                              //nolint:depguard,nolintlint
)

// Functions return following wrapped errors: ErrInvalidID, ErrInvalidFieldData, ErrEmptyField.

func idFromString(id string) (*uuid.UUID, error) {
	res, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", packageErrors.ErrInvalidID, err)
	}
	return &res, nil
}

func fromInternalEvent(event *types.Event) *pb.Event {
	return &pb.Event{
		Id:   event.ID.String(),
		Data: fromInternalEventData(&event.EventData),
	}
}

func fromInternalEventData(data *types.EventData) *pb.EventData {
	var remindIn *durationpb.Duration

	if data.RemindIn > 0 {
		remindIn = durationpb.New(data.RemindIn)
	}

	return &pb.EventData{
		Title:       data.Title,
		Datetime:    timestamppb.New(data.Datetime),
		Duration:    durationpb.New(data.Duration),
		Description: data.Description,
		UserId:      data.UserID,
		RemindIn:    remindIn,
	}
}
