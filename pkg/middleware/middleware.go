package middleware

import (
	"errors"
	"net/http"
	"time"

	"aether-node/internal/metrics"
	"aether-node/pkg/logger"
	"aether-node/pkg/response"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// JWT Error codes for granular error handling
const (
	JWTErrorCodeExpired           = "TOKEN_EXPIRED"
	JWTErrorCodeMalformed         = "TOKEN_MALFORMED"
	JWTErrorCodeSignatureInvalid  = "TOKEN_SIGNATURE_INVALID"
	JWTErrorCodeInvalid           = "TOKEN_INVALID"
	JWTErrorCodeMissing           = "TOKEN_MISSING"
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

// JWTAuthErrorHandler returns an Echo error handler that maps JWT errors to specific error codes.
// This provides granular error responses for expired, malformed, invalid signature, and other JWT errors.
func JWTAuthErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var (
			code    int
			errCode string
			message string
		)

		he, ok := err.(*echo.HTTPError)
		if ok {
			if he.Internal != nil {
				err = he.Internal
			}
		}

		// Check for our custom JWTAuthError first
		if jwtErr, ok := err.(*JWTAuthError); ok {
			err = jwtErr.Err // Unwrap to check jwt-specific errors
			errCode = jwtErr.ErrCode
			message = jwtErr.ErrReason
			code = http.StatusUnauthorized
		}

		switch {
		case err == nil:
			// No error, shouldn't happen but handle anyway
			code = http.StatusInternalServerError
			errCode = JWTErrorCodeInvalid
			message = "Unknown JWT error"

		case errorsIs(err, jwt.ErrTokenExpired):
			if errCode == "" {
				errCode = JWTErrorCodeExpired
			}
			if message == "" {
				message = "Access token has expired"
			}
			code = http.StatusUnauthorized

		case errorsIs(err, jwt.ErrTokenMalformed):
			if errCode == "" {
				errCode = JWTErrorCodeMalformed
			}
			if message == "" {
				message = "Malformed access token"
			}
			code = http.StatusUnauthorized

		case errorsIs(err, jwt.ErrTokenSignatureInvalid):
			if errCode == "" {
				errCode = JWTErrorCodeSignatureInvalid
			}
			if message == "" {
				message = "Invalid token signature"
			}
			code = http.StatusUnauthorized

		case errorsIs(err, jwt.ErrTokenInvalidClaims):
			if errCode == "" {
				errCode = JWTErrorCodeInvalid
			}
			if message == "" {
				message = "Invalid token claims"
			}
			code = http.StatusUnauthorized

		case errorsIs(err, jwt.ErrTokenNotValidYet):
			if errCode == "" {
				errCode = JWTErrorCodeInvalid
			}
			if message == "" {
				message = "Token not valid yet"
			}
			code = http.StatusUnauthorized

		case errorsIs(err, echo.ErrUnauthorized):
			if errCode == "" {
				errCode = JWTErrorCodeInvalid
			}
			if message == "" {
				message = "Unauthorized"
			}
			code = http.StatusUnauthorized

		default:
			if errCode == "" {
				errCode = JWTErrorCodeInvalid
			}
			if message == "" {
				message = "Invalid or expired access token"
			}
			code = http.StatusUnauthorized
		}

		// Send JSON error response (not string)
		if c.Request().Method == http.MethodHead {
			c.NoContent(code)
		} else {
			response.Error(c, code, errCode, message)
		}
	}
}

// JWTAuthErrorHandlerFunc is an adapter for JWT middleware's ErrorHandler signature.
// It maps JWT errors to specific error codes and returns an error that will be handled
// by the global HTTPErrorHandler.
func JWTAuthErrorHandlerFunc(err error) error {
	// We need to store the error and handle it in HTTPErrorHandler
	// Since JWT middleware calls ErrorHandler and then might call c.JSON(),
	// we need to use a custom error type that HTTPErrorHandler can recognize
	return &JWTAuthError{
		Err:       err,
		ErrCode:   getJWTErrCode(err),
		ErrReason: getJWTErrReason(err),
	}
}

// JWTAuthErrorHandlerWithContext is the ErrorHandlerWithContext for JWT middleware.
// It directly writes the error response with specific error codes.
func JWTAuthErrorHandlerWithContext(err error, c echo.Context) error {
	if c.Response().Committed {
		return nil
	}

	var (
		code    int
		errCode string
		message string
	)

	// Unwrap HTTPError if present
	if he, ok := err.(*echo.HTTPError); ok {
		if he.Internal != nil {
			err = he.Internal
		}
	}

	// Check for our custom JWTAuthError first
	if jwtErr, ok := err.(*JWTAuthError); ok {
		err = jwtErr.Err // Unwrap to check jwt-specific errors
		errCode = jwtErr.ErrCode
		message = jwtErr.ErrReason
		code = http.StatusUnauthorized
	}

	// Direct comparison - if matched, we're done
	if err == jwt.ErrTokenExpired {
		errCode = "TOKEN_EXPIRED"
		message = "Direct match: token is expired"
		code = http.StatusUnauthorized
		if c.Request().Method == http.MethodHead {
			return c.NoContent(code)
		}
		return response.Error(c, code, errCode, message)
	}

	// DEBUG: print error type and message
	var errType string
	switch err.(type) {
	case *echo.HTTPError:
		errType = "*echo.HTTPError"
	case error:
		errType = "plain error: " + err.Error()
	default:
		errType = "unknown"
	}
	println("DEBUG errType:", errType)
	println("DEBUG errMsg:", err.Error())
	if err == jwt.ErrTokenExpired {
		println("DEBUG: Direct match with jwt.ErrTokenExpired!")
	}

	switch {
	case err == nil:
		code = http.StatusInternalServerError
		errCode = JWTErrorCodeInvalid
		message = "Unknown JWT error"

	case errorsIs(err, jwt.ErrTokenExpired):
		if errCode == "" {
			errCode = JWTErrorCodeExpired
		}
		if message == "" {
			message = "Access token has expired"
		}
		code = http.StatusUnauthorized

	case errorsIs(err, jwt.ErrTokenMalformed):
		if errCode == "" {
			errCode = JWTErrorCodeMalformed
		}
		if message == "" {
			message = "Malformed access token"
		}
		code = http.StatusUnauthorized

	case errorsIs(err, jwt.ErrTokenSignatureInvalid):
		if errCode == "" {
			errCode = JWTErrorCodeSignatureInvalid
		}
		if message == "" {
			message = "Invalid token signature"
		}
		code = http.StatusUnauthorized

	case errorsIs(err, jwt.ErrTokenInvalidClaims):
		if errCode == "" {
			errCode = JWTErrorCodeInvalid
		}
		if message == "" {
			message = "Invalid token claims"
		}
		code = http.StatusUnauthorized

	case errorsIs(err, jwt.ErrTokenNotValidYet):
		if errCode == "" {
			errCode = JWTErrorCodeInvalid
		}
		if message == "" {
			message = "Token not valid yet"
		}
		code = http.StatusUnauthorized

	case errorsIs(err, echo.ErrUnauthorized):
		if errCode == "" {
			errCode = JWTErrorCodeInvalid
		}
		if message == "" {
			message = "Unauthorized"
		}
		code = http.StatusUnauthorized

	default:
		if errCode == "" {
			errCode = JWTErrorCodeInvalid
		}
		if message == "" {
			message = "Invalid or expired access token"
		}
		code = http.StatusUnauthorized
	}

	// Send JSON error response directly
	if c.Request().Method == http.MethodHead {
		return c.NoContent(code)
	}
	return response.Error(c, code, errCode, message)
}

// errorsIs is a wrapper for errors.Is that handles nil errors safely
func errorsIs(err, target error) bool {
	if err == nil || target == nil {
		return false
	}
	return errors.Is(err, target)
}

// JWTAuthError is a custom error that carries JWT error metadata
type JWTAuthError struct {
	Err       error
	ErrCode   string
	ErrReason string
}

func (e *JWTAuthError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "jwt error"
}

func (e *JWTAuthError) Unwrap() error {
	return e.Err
}

// getJWTErrCode returns the appropriate error code for a JWT error
func getJWTErrCode(err error) string {
	if err == nil {
		return JWTErrorCodeInvalid
	}
	switch {
	case errorsIs(err, jwt.ErrTokenExpired):
		return JWTErrorCodeExpired
	case errorsIs(err, jwt.ErrTokenMalformed):
		return JWTErrorCodeMalformed
	case errorsIs(err, jwt.ErrTokenSignatureInvalid):
		return JWTErrorCodeSignatureInvalid
	default:
		return JWTErrorCodeInvalid
	}
}

// getJWTErrReason returns the appropriate error reason/message for a JWT error
func getJWTErrReason(err error) string {
	if err == nil {
		return "Invalid or expired access token"
	}
	switch {
	case errorsIs(err, jwt.ErrTokenExpired):
		return "Access token has expired"
	case errorsIs(err, jwt.ErrTokenMalformed):
		return "Malformed access token"
	case errorsIs(err, jwt.ErrTokenSignatureInvalid):
		return "Invalid token signature"
	default:
		return "Invalid or expired access token"
	}
}
