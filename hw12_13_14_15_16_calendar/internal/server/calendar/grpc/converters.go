package grpc

import (
	"time"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"          //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/durationpb"                         //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/timestamppb"                        //nolint:depguard,nolintlint
)

// Functions return following wrapped errors: ErrInvalidID, ErrInvalidFieldData, ErrEmptyField.

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

func setDesctription(description string) *string {
	desc := description
	if desc != "" {
		return &desc
	}
	return nil
}

func setDuration(reqDuration *durationpb.Duration) *time.Duration {
	if reqDuration == nil {
		return nil
	}
	res := reqDuration.AsDuration()
	return &res
}

func setTime(reqTime *timestamppb.Timestamp) time.Time {
	if reqTime == nil {
		return time.Time{}
	}
	res := reqTime.AsTime()
	return res
}
