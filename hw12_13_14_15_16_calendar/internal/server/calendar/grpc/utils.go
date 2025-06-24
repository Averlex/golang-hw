package grpc

import (
	"context"
	"errors"
	"log/slog"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1"       //nolint:depguard,nolintlint
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"google.golang.org/grpc/codes"                                                    //nolint:depguard,nolintlint
	"google.golang.org/grpc/status"                                                   //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/anypb"                                    //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/wrapperspb"                               //nolint:depguard,nolintlint
)

func (s *Server) wrapError(ctx context.Context, err error) *pb.Status {
	var st *pb.Status
	if err == nil {
		return &pb.Status{
			Code:    int32(codes.OK),
			Message: "Success",
		}
	}

	// Preparing error details.
	anyDetail, wrapErr := anypb.New(wrapperspb.String(err.Error()))
	if wrapErr != nil {
		s.l.Error(ctx, "failed to wrap orignal error", slog.String("err", err.Error()))
		return &pb.Status{
			Code:    int32(codes.Internal),
			Message: status.Errorf(codes.Internal, "Failed to wrap orignal error").Error(),
		}
	}

	// Processing different kinds of errors.
	switch {
	// Crucial errors.
	case errors.Is(err, projectErrors.ErrStorageFull):
		s.l.Error(ctx, "storage is full", slog.String("err", err.Error()))
		st = &pb.Status{
			Code:    int32(codes.ResourceExhausted),
			Message: status.Errorf(codes.ResourceExhausted, "Storage is full").Error(),
			Details: []*anypb.Any{anyDetail},
		}
	case errors.Is(err, projectErrors.ErrInconsistentState):
		st = &pb.Status{
			Code:    int32(codes.Internal),
			Message: status.Errorf(codes.Internal, "Unexpected internal error occurred").Error(),
			Details: []*anypb.Any{anyDetail},
		}
	// This one means the user can try again.
	// We should warn the dev about this anyway - it might be a connection error, query error or something else,
	// which might involve a developer intervention.
	case errors.Is(err, projectErrors.ErrRetriesExceeded):
		s.l.Warn(ctx, "retries exceeded", slog.String("err", err.Error()))
		st = &pb.Status{
			Code:    int32(codes.ResourceExhausted),
			Message: status.Errorf(codes.ResourceExhausted, "Retries exceeded. Please, try again later").Error(),
			Details: []*anypb.Any{anyDetail},
		}
	// Business errors.
	case errors.Is(err, projectErrors.ErrEmptyField), errors.Is(err, projectErrors.ErrInvalidFieldData):
		st = &pb.Status{
			Code:    int32(codes.InvalidArgument),
			Message: status.Errorf(codes.InvalidArgument, "Provided data is invalid").Error(),
			Details: []*anypb.Any{anyDetail},
		}
	case errors.Is(err, projectErrors.ErrNoData):
		st = &pb.Status{
			Code:    int32(codes.InvalidArgument),
			Message: status.Errorf(codes.InvalidArgument, "Request received no data").Error(),
			Details: []*anypb.Any{anyDetail},
		}
	case errors.Is(err, projectErrors.ErrEventNotFound):
		st = &pb.Status{
			Code:    int32(codes.NotFound),
			Message: status.Errorf(codes.NotFound, "Requested event was not found").Error(),
			Details: []*anypb.Any{anyDetail},
		}
	case errors.Is(err, projectErrors.ErrDateBusy):
		st = &pb.Status{
			Code:    int32(codes.AlreadyExists),
			Message: status.Errorf(codes.AlreadyExists, "Requested event date is already busy").Error(),
			Details: []*anypb.Any{anyDetail},
		}
	case errors.Is(err, projectErrors.ErrPermissionDenied):
		st = &pb.Status{
			Code:    int32(codes.PermissionDenied),
			Message: status.Errorf(codes.PermissionDenied, "Cannot modify another user's event").Error(),
			Details: []*anypb.Any{anyDetail},
		}
	}

	return st
}
