//nolint:revive
package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/gin-gonic/gin"                                                        //nolint:depguard,nolintlint
)

// Server represents an HTTP server with a gin engine.
type Server struct {
	mu              sync.RWMutex
	a               Application
	l               Logger
	shutdownTimeout time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	engine          *gin.Engine
	srv             *http.Server
	addr            string
}

// NewServer creates a new HTTP server. The function performs validation of the input parameters.
// If no error occurs, it returns *Server, nil and nil, error otherwise.
func NewServer(logger Logger, app Application, config map[string]any) (*Server, error) {
	// Args validation.
	missing := make([]string, 0)
	if logger == nil {
		missing = append(missing, "logger")
	}
	if app == nil {
		missing = append(missing, "app")
	}
	if config == nil {
		missing = append(missing, "config")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: some of the required parameters are missing: args=%v",
			projectErrors.ErrAppInitFailed, missing)
	}

	// Field types validation.
	missing, wrongType := validateFields(config, expectedFields)
	if len(missing) > 0 || len(wrongType) > 0 {
		return nil, fmt.Errorf("%w: missing=%v invalid_type=%v",
			projectErrors.ErrCorruptedConfig, missing, wrongType)
	}

	// Extract from config an normalize the value.
	host, _ := config["host"].(string)
	port, _ := config["port"].(string)
	shutdownTimeout, _ := config["shutdown_timeout"].(time.Duration)
	readTimeout, _ := config["read_timeout"].(time.Duration)
	writeTimeout, _ := config["write_timeout"].(time.Duration)
	idleTimeout, _ := config["idle_timeout"].(time.Duration)
	shutdownTimeout = max(0, shutdownTimeout)

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
		readTimeout:     readTimeout,
		writeTimeout:    writeTimeout,
		idleTimeout:     idleTimeout,
		addr:            fmt.Sprintf("%s:%s", host, port),
	}, nil
}

// Start starts the HTTP server. Start blocks the calling goroutine until the error returns.
func (s *Server) Start(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(s.requestContextMiddleware(ctx))
	engine.Use(s.loggingMiddleware())

	engine.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/hello")
	})

	s.mu.Lock()
	s.engine = engine
	addr := s.addr
	s.srv = &http.Server{
		Addr:         addr,
		Handler:      engine,
		WriteTimeout: s.writeTimeout,
		ReadTimeout:  s.readTimeout,
		IdleTimeout:  s.idleTimeout,
	}
	s.mu.Unlock()

	s.l.Info(ctx, "starting HTTP server", slog.String("addr", addr))

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
