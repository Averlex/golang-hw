package http

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin" //nolint:depguard,nolintlint
	"github.com/google/uuid"   //nolint:depguard,nolintlint
)

// requestDataKey is a key for storing request data in the context.
var requestDataKey = "request_id"

// RequestData represents a data structure for storing request data.
type RequestData struct {
	ClientIP   string
	Method     string
	Path       string
	Proto      string
	UserAgent  string
	StartTime  time.Time
	StatusCode int
}

// requestContextMiddleware adds a request ID to the context service context.
// It is placed in the gin context to provide an access to other middleware and service layers.
//
// It is meant to be used as a first middleware in the chain.
func (s *Server) requestContextMiddleware(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		//nolint:staticcheck,revive
		ctx := context.WithValue(ctx, requestDataKey, slog.String(requestDataKey, requestID))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// loggingMiddleware logs requests, along with the execution time and status code.
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		startTime := time.Now()

		c.Next()

		s.l.Info(ctx, "request processed",
			slog.String("client_ip", c.ClientIP()),
			slog.Time("start_time", time.Now()),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("proto", c.Request.Proto),
			slog.Int("status_code", c.Writer.Status()),
			slog.Duration("latency", time.Since(startTime)),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}
