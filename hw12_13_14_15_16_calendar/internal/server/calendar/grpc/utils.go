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

	var msg error
	// Processing different kinds of errors.
	switch {
	// Crucial errors.
	case errors.Is(err, projectErrors.ErrStorageFull):
		s.l.Error(ctx, "storage is full", slog.String("err", err.Error()))
		msg = status.Errorf(codes.ResourceExhausted, "Storage is full")
	case errors.Is(err, projectErrors.ErrInconsistentState):
		msg = status.Errorf(codes.Internal, "Unexpected internal error occurred")
	// This one means the user can try again.
	// We should warn the dev about this anyway - it might be a connection error, query error or something else,
	// which might involve a developer intervention.
	case errors.Is(err, projectErrors.ErrRetriesExceeded):
		s.l.Warn(ctx, "retries exceeded", slog.String("err", err.Error()))
		msg = status.Errorf(codes.ResourceExhausted, "Retries exceeded. Please, try again later")
	// Business errors.
	case errors.Is(err, projectErrors.ErrEmptyField), errors.Is(err, projectErrors.ErrInvalidFieldData):
		msg = status.Errorf(codes.InvalidArgument, "Provided data is invalid")
	case errors.Is(err, projectErrors.ErrNoData):
		msg = status.Errorf(codes.InvalidArgument, "Request received no data")
	case errors.Is(err, projectErrors.ErrEventNotFound):
		msg = status.Errorf(codes.NotFound, "Requested event was not found")
	case errors.Is(err, projectErrors.ErrDateBusy):
		msg = status.Errorf(codes.AlreadyExists, "Requested event date is already busy")
	case errors.Is(err, projectErrors.ErrPermissionDenied):
		msg = status.Errorf(codes.PermissionDenied, "Cannot modify another user's event")
	default:
		s.l.Error(ctx, "unknown error received", slog.String("err", err.Error()))
		msg = status.Errorf(codes.Internal, "Unexpected internal error occurred")
	}

	return &pb.Status{
		Code:    int32(status.Code(msg)), //nolint:gosec
		Message: msg.Error(),
		Details: anyDetail,
	}
}
