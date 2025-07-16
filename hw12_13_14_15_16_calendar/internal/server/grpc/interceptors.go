package grpc

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"        //nolint:depguard,nolintlint
	"google.golang.org/grpc"        //nolint:depguard,nolintlint
	"google.golang.org/grpc/peer"   //nolint:depguard,nolintlint
	"google.golang.org/grpc/status" //nolint:depguard,nolintlint
)

// requestDataKey is a key for storing request data in the context.
var requestDataKey = "grpc_request_id"

// RequestData represents a data structure for storing request data.
type RequestData struct {
	ClientIP   string
	Method     string
	StartTime  time.Time
	StatusCode int
}

// requestContextUnaryInterceptor adds a gRPC request ID to the context service context.
// It is placed in the gin context to provide an access to other middleware and service layers.
//
// It is meant to be used as a first middleware in the chain.
func (s *Server) requestContextUnaryInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	requestID := uuid.New().String()
	//nolint:staticcheck,revive
	childCtx := context.WithValue(ctx, requestDataKey, slog.String(requestDataKey, requestID))

	resp, err := handler(childCtx, req)

	return resp, err
}

// loggingUnaryInterceptor logs requests, along with the execution time and status code.
func (s *Server) loggingUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	startTime := time.Now()

	// Extracting client IP.
	peer, _ := peer.FromContext(ctx)
	clientIP := "unknown"
	if peer != nil {
		clientIP = peer.Addr.String()
	}

	resp, err := handler(ctx, req)

	// Exctracting status code.
	statusCode := "OK"
	if sCode, ok := status.FromError(err); ok {
		statusCode = sCode.Code().String()
	}

	s.l.Info(ctx, "gRPC call",
		slog.String("client_ip", clientIP),
		slog.Time("start_time", time.Now()),
		slog.String("method", info.FullMethod),
		slog.String("status_code", statusCode),
		slog.Duration("latency", time.Since(startTime)),
	)

	return resp, err
}
