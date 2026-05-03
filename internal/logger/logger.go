package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// CtxKey type for context keys
type CtxKey string

const (
	RequestIDKey CtxKey = "request_id"
	UserIDKey    CtxKey = "user_id"
)

// Logger wraps zerolog.Logger with additional context
type Logger struct {
	zerolog.Logger
}

// New creates a new structured logger
func New(env string) *Logger {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var z zerolog.Logger
	if env == "production" {
		z = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	} else {
		z = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Caller().
			Logger().
			Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339Nano})
	}

	return &Logger{Logger: z}
}

// WithRequestID adds request_id to context
func (l *Logger) WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserID adds user_id to context
func (l *Logger) WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithCtx returns a logger with context fields (request_id, user_id)
func (l *Logger) WithCtx(ctx context.Context) *Logger {
	reqID, _ := ctx.Value(RequestIDKey).(string)
	userID, _ := ctx.Value(UserIDKey).(string)

	event := l.With()
	if reqID != "" {
		event = event.Str("request_id", reqID)
	}
	if userID != "" {
		event = event.Str("user_id", userID)
	}

	return &Logger{Logger: event.Logger()}
}

// InfoCtx logs info with context
func (l *Logger) InfoCtx(ctx context.Context) *zerolog.Event {
	return l.WithCtx(ctx).Info()
}

// ErrorCtx logs error with context
func (l *Logger) ErrorCtx(ctx context.Context) *zerolog.Event {
	return l.WithCtx(ctx).Error()
}

// WarnCtx logs warn with context
func (l *Logger) WarnCtx(ctx context.Context) *zerolog.Event {
	return l.WithCtx(ctx).Warn()
}

// DebugCtx logs debug with context
func (l *Logger) DebugCtx(ctx context.Context) *zerolog.Event {
	return l.WithCtx(ctx).Debug()
}

// FatalCtx logs fatal with context
func (l *Logger) FatalCtx(ctx context.Context) *zerolog.Event {
	return l.WithCtx(ctx).Fatal()
}

// LogHTTPRequest logs HTTP request with structured fields
func (l *Logger) LogHTTPRequest(method, path string, status int, duration time.Duration, requestID, userID string) {
	l.Info().
		Str("request_id", requestID).
		Str("user_id", userID).
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Dur("duration_ms", duration).
		Msg("http_request")
}

// LogSSEConnection logs SSE connection events
func (l *Logger) LogSSEConnection(event, requestID, deviceSN string, activeConnections int) {
	l.Info().
		Str("event", event).
		Str("request_id", requestID).
		Str("device_sn", deviceSN).
		Int("active_connections", activeConnections).
		Msg("sse_connection")
}

// LogInfluxDBQuery logs InfluxDB query with duration and status
func (l *Logger) LogInfluxDBQuery(operation, requestID string, duration time.Duration, success bool, err error) {
	event := l.Info()
	if !success {
		event = l.Error()
	}
	event.
		Str("request_id", requestID).
		Str("operation", operation).
		Dur("duration_ms", duration).
		Bool("success", success).
		Err(err).
		Msg("influxdb_query")
}
