// Package grpc provides a gRPC server implementation.
package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1"            //nolint:depguard,nolintlint
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
	"google.golang.org/grpc"                                                               //nolint:depguard,nolintlint
	"google.golang.org/grpc/reflection"                                                    //nolint:depguard,nolintlint
)

// Server represents a gRPC server.
type Server struct {
	pb.UnimplementedCalendarServiceServer

	a Application
	l Logger

	server *grpc.Server
	lis    net.Listener
	mu     sync.Mutex

	addr            string
	shutdownTimeout time.Duration
}

// NewServer creates a new gRPC server. The function performs validation of the input parameters.
// If no error occurs, it returns *Server, nil and nil, error otherwise.
func NewServer(logger Logger, app Application, config map[string]any) (*Server, error) {
	// Args validation.
	missing := make([]string, 0)
	if logger == nil {
		missing = append(missing, "logger")
	}
	if app == nil {
		missing = append(missing, "a")
	}
	if config == nil {
		missing = append(missing, "config")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: some of the required parameters are missing: args=%v",
			projectErrors.ErrServerInitFailed, missing)
	}

	// Field types validation.
	missing, wrongType := validateFields(config, expectedFields)
	if len(missing) > 0 || len(wrongType) > 0 {
		return nil, fmt.Errorf("%w: missing=%v invalid_type=%v",
			projectErrors.ErrCorruptedConfig, missing, wrongType)
	}

	/// Extract from config an normalize the value.
	host, _ := config["host"].(string)
	port, _ := config["port"].(string)
	shutdownTimeout, _ := config["shutdown_timeout"].(time.Duration)

	invalidValues := make([]string, 0)
	if host == "" {
		invalidValues = append(invalidValues, "host")
	}
	if port == "" {
		invalidValues = append(invalidValues, "port")
	}
	if len(invalidValues) > 0 {
		return nil, fmt.Errorf("%w: invalid values=%v", projectErrors.ErrCorruptedConfig, invalidValues)
	}

	return &Server{
		a:               app,
		l:               logger,
		shutdownTimeout: shutdownTimeout,
		addr:            fmt.Sprintf("%s:%s", host, port),
	}, nil
}

// Start starts the gRPC server. Start blocks the calling goroutine until the error returns.
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr := s.addr
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen tcp at %v: %w", addr, err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.requestContextUnaryInterceptor,
			s.loggingUnaryInterceptor,
		),
	)

	pb.RegisterCalendarServiceServer(grpcServer, s)

	s.server = grpcServer
	s.lis = lis
	reflection.Register(s.server)

	// Avoiding server start.
	select {
	case <-ctx.Done():
		return lis.Close()
	default:
	}

	s.l.Info(ctx, "starting gRPC server", slog.String("addr", addr))
	return grpcServer.Serve(lis)
}

// Stop gracefully shuts down the gRPC server.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()

	server := s.server
	lis := s.lis

	s.mu.Unlock()

	if server == nil {
		s.l.Warn(ctx, "gRPC server is not running")
		return nil
	}

	if lis == nil {
		return fmt.Errorf("gRPC listener is not running")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.l.Info(ctx, "gRPC server stopped successfully")
	case <-shutdownCtx.Done():
		server.Stop()
		s.l.Warn(ctx, "gRPC server stoppped forcefully", slog.String("error", shutdownCtx.Err().Error()))
	}

	return nil
}
