package middleware

import (
	"time"

	"aether-node/internal/metrics"
	"aether-node/pkg/logger"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

const (
	// RequestIDHeader is the header name for request correlation ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDContextKey is the context key for request ID
	RequestIDContextKey = "request_id"
)

// RequestLogger returns a middleware that logs HTTP requests with structured logging.
// It adds a correlation ID (X-Request-ID header) to each request for tracing.
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Get or generate request ID
			requestID := req.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Set request ID in response header
			res.Header().Set(RequestIDHeader, requestID)

			// Store in context for use in handlers
			c.Set(RequestIDContextKey, requestID)

			// Create request-scoped logger
			log := logger.WithRequestID(requestID)

			// Start timer
			start := time.Now()

			// Process request
			err := next(c)

			// Calculate latency
			latency := time.Since(start)

			// Record Prometheus metrics
			duration := latency.Seconds()
			status := res.Status
			metrics.RecordHTTPRequest(req.Method, req.URL.Path, status, duration)

			// Build log event
			event := log.Info()
			if err != nil {
				event = log.Error()
			}

			// Add request details
			event.
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", status).
				Dur("latency", latency).
				Str("remote_ip", c.RealIP()).
				Int64("bytes_out", res.Size).
				Msg("HTTP request")

			return err
		}
	}
}

// GetRequestID extracts the request ID from the Echo context
func GetRequestID(c echo.Context) string {
	if id, ok := c.Get(RequestIDContextKey).(string); ok {
		return id
	}
	return ""
}

// ComponentLogger returns a logger with a component field
func ComponentLogger(component string) zerolog.Logger {
	return logger.WithComponent(component)
}
