package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

// Init initializes the global logger.
// level: "debug", "info", "warn", "error" (default: "info")
// jsonOutput: true for JSON format (production), false for console (development)
func Init(level string, jsonOutput bool) {
	// Set log level
	l := parseLevel(level)
	zerolog.SetGlobalLevel(l)

	var output io.Writer = os.Stdout
	if !jsonOutput {
		// Pretty console output for development
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	log = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()
}

// parseLevel converts string to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Get returns the global logger instance
func Get() *zerolog.Logger {
	return &log
}

// Debug logs a debug message
func Debug() *zerolog.Event {
	return log.Debug()
}

// Info logs an info message
func Info() *zerolog.Event {
	return log.Info()
}

// Warn logs a warning message
func Warn() *zerolog.Event {
	return log.Warn()
}

// Error logs an error message
func Error() *zerolog.Event {
	return log.Error()
}

// Fatal logs a fatal message and exits
func Fatal() *zerolog.Event {
	return log.Fatal()
}

// WithComponent returns a logger with a "component" field
func WithComponent(component string) zerolog.Logger {
	return log.With().Str("component", component).Logger()
}

// WithRequestID returns a logger with a "request_id" field
func WithRequestID(requestID string) zerolog.Logger {
	return log.With().Str("request_id", requestID).Logger()
}

// WithUserGUID returns a logger with a "user_guid" field
func WithUserGUID(userGUID string) zerolog.Logger {
	return log.With().Str("user_guid", userGUID).Logger()
}
