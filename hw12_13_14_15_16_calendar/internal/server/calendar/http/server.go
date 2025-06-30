//nolint:revive
package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	pb "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/api/calendar/v1"       //nolint:depguard,nolintlint
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/gin-gonic/gin"                                                        //nolint:depguard,nolintlint
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"                               //nolint:depguard,nolintlint
	"google.golang.org/grpc"                                                          //nolint:depguard,nolintlint
	"google.golang.org/grpc/credentials/insecure"                                     //nolint:depguard,nolintlint
)

// Server represents an HTTP server with a gin engine.
type Server struct {
	mu sync.RWMutex
	a  Application
	l  Logger

	shutdownTimeout time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration

	engine *gin.Engine
	srv    *http.Server

	httpAddr string
	grpcAddr string
}

// NewServer creates a new HTTP server. The function performs validation of the input parameters.
// If no error occurs, it returns *Server, nil and nil, error otherwise.
func NewServer(logger Logger, app Application, httpConfig map[string]any, grpcConfig map[string]any) (*Server, error) {
	// Args validation.
	missing := make([]string, 0)
	if logger == nil {
		missing = append(missing, "logger")
	}
	if app == nil {
		missing = append(missing, "app")
	}
	if httpConfig == nil {
		missing = append(missing, "config")
	}
	if grpcConfig == nil {
		missing = append(missing, "grpc_config")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: some of the required parameters are missing: args=%v",
			projectErrors.ErrServerInitFailed, missing)
	}

	// Field types validation.
	missing, wrongType := validateFields(httpConfig, expectedHTTPFields)         // HTTP config.
	grpcMissing, grpcWrongType := validateFields(grpcConfig, expectedGRPCFields) // gRPC config.
	missing = append(missing, grpcMissing...)
	wrongType = append(wrongType, grpcWrongType...)
	if len(missing) > 0 || len(wrongType) > 0 {
		return nil, fmt.Errorf("%w: missing=%v invalid_type=%v",
			projectErrors.ErrCorruptedConfig, missing, wrongType)
	}

	// Extract from config an normalize the value.
	httpHost, _ := httpConfig["host"].(string)
	httpPort, _ := httpConfig["port"].(string)
	shutdownTimeout, _ := httpConfig["shutdown_timeout"].(time.Duration)
	readTimeout, _ := httpConfig["read_timeout"].(time.Duration)
	writeTimeout, _ := httpConfig["write_timeout"].(time.Duration)
	idleTimeout, _ := httpConfig["idle_timeout"].(time.Duration)
	shutdownTimeout = max(0, shutdownTimeout)
	grpcHost, _ := grpcConfig["host"].(string)
	grpcPort, _ := grpcConfig["port"].(string)

	invalidValues := make([]string, 0)
	if httpHost == "" {
		invalidValues = append(invalidValues, "host")
	}
	if httpPort == "" {
		invalidValues = append(invalidValues, "port")
	}
	if grpcHost == "" {
		invalidValues = append(invalidValues, "grpc_host")
	}
	if grpcPort == "" {
		invalidValues = append(invalidValues, "grpc_port")
	}
	if len(invalidValues) > 0 {
		return nil, fmt.Errorf("%w: invalid values=%v", projectErrors.ErrCorruptedConfig, invalidValues)
	}

	return &Server{
		a:               app,
		l:               logger,
		shutdownTimeout: shutdownTimeout,
		readTimeout:     readTimeout,
		writeTimeout:    writeTimeout,
		idleTimeout:     idleTimeout,
		httpAddr:        fmt.Sprintf("%s:%s", httpHost, httpPort),
		grpcAddr:        fmt.Sprintf("%s:%s", grpcHost, grpcPort),
	}, nil
}

// Start starts the HTTP server. Start blocks the calling goroutine until the error returns.
func (s *Server) Start(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(s.requestContextMiddleware(ctx))
	engine.Use(s.loggingMiddleware())

	// Test endpoints.
	engine.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/hello")
	})

	s.mu.Lock()
	// Inititializing gRPC gateway and its routing.
	grpcAddr := s.grpcAddr
	gwHandler, err := s.initGRPCGateway(ctx, grpcAddr)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to init gRPC gateway: %w", err)
	}

	// Register protobuf endpoints. Non-blocking.
	engine.Any("/v1/*any", func(c *gin.Context) {
		if gwHandler == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "gRPC-Gateway handler not initialized",
			})
			return
		}
		gwHandler.ServeHTTP(c.Writer, c.Request)
	})

	s.engine = engine
	httpAddr := s.httpAddr
	s.srv = &http.Server{
		Addr:         httpAddr,
		Handler:      engine,
		WriteTimeout: s.writeTimeout,
		ReadTimeout:  s.readTimeout,
		IdleTimeout:  s.idleTimeout,
	}

	s.mu.Unlock()

	s.l.Info(ctx, "starting HTTP server", slog.String("addr", httpAddr))

	return s.srv.ListenAndServe()
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.RLock()
	srv := s.srv
	s.mu.RUnlock()

	if srv == nil {
		s.l.Warn(ctx, "HTTP server is not running")
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		s.l.Error(ctx, "HTTP server graceful shutdown", slog.Any("error", err))
		return err
	}

	s.l.Info(ctx, "HTTP server stopped successfully")
	return nil
}

func (s *Server) initGRPCGateway(ctx context.Context, grpcEndpoint string) (http.Handler, error) {
	// Register the gRPC server endpoint.
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	err := pb.RegisterCalendarServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to register gRPC handler: %w", err)
	}

	return mux, nil
}
