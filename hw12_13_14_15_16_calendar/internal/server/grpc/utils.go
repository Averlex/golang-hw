package grpc

import (
	"context"
	"errors"
	"log/slog"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
	"google.golang.org/grpc"                                                               //nolint:depguard,nolintlint
	"google.golang.org/grpc/codes"                                                         //nolint:depguard,nolintlint
	"google.golang.org/grpc/metadata"                                                      //nolint:depguard,nolintlint
	"google.golang.org/grpc/status"                                                        //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/anypb"                                         //nolint:depguard,nolintlint
	"google.golang.org/protobuf/types/known/wrapperspb"                                    //nolint:depguard,nolintlint
)

func (s *Server) wrapError(ctx context.Context, err error) *status.Status {
	if err == nil {
		return status.New(codes.OK, "Success")
	}

	// Preparing error details.
	anyDetail, wrapErr := anypb.New(wrapperspb.String(err.Error()))
	if wrapErr != nil {
		s.l.Error(ctx, "failed to wrap orignal error", slog.String("err", err.Error()))
		return status.New(codes.Internal, "Failed to wrap orignal error")
	}

	var st *status.Status
	// Processing different kinds of errors.
	switch {
	// Crucial errors.
	case errors.Is(err, projectErrors.ErrStorageFull):
		s.l.Error(ctx, "storage is full", slog.String("err", err.Error()))
		st = status.New(codes.ResourceExhausted, "Storage is full")
	case errors.Is(err, projectErrors.ErrInconsistentState):
		st = status.New(codes.Internal, "Unexpected internal error occurred")
	// This one means the user can try again.
	// We should warn the dev about this anyway - it might be a connection error, query error or something else,
	// which might involve a developer intervention.
	case errors.Is(err, projectErrors.ErrRetriesExceeded):
		s.l.Warn(ctx, "retries exceeded", slog.String("err", err.Error()))
		st = status.New(codes.ResourceExhausted, "Retries exceeded. Please, try again later")
	// Business errors.
	case errors.Is(err, projectErrors.ErrEmptyField), errors.Is(err, projectErrors.ErrInvalidFieldData):
		st = status.New(codes.InvalidArgument, "Provided data is invalid")
	case errors.Is(err, projectErrors.ErrNoData):
		st = status.New(codes.InvalidArgument, "Request received no data")
	case errors.Is(err, projectErrors.ErrEventNotFound):
		st = status.New(codes.NotFound, "Requested event was not found")
	case errors.Is(err, projectErrors.ErrDateBusy):
		st = status.New(codes.AlreadyExists, "Requested event date is already busy")
	case errors.Is(err, projectErrors.ErrPermissionDenied):
		st = status.New(codes.PermissionDenied, "Cannot modify another user's event")
	default:
		s.l.Error(ctx, "unknown error received", slog.String("err", err.Error()))
		st = status.New(codes.Internal, "Unexpected internal error occurred")
	}

	resSt, err := st.WithDetails(anyDetail)
	if err != nil {
		s.l.Error(ctx, "failed to add error details", slog.String("err", err.Error()))
		return st
	}
	st = resSt

	return st
}

// handleError wraps the error and sets gRPC headers.
func (s *Server) handleError(ctx context.Context, err error) *status.Status {
	st := s.wrapError(ctx, err)
	_ = grpc.SetHeader(ctx, metadata.MD{})
	return st
}
