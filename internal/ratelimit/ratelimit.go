package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Config holds rate limiter configuration
type Config struct {
	// Skipper defines a function to skip this middleware when returned true.
	Skipper middleware.Skipper

	// IdentifierExtractor extracts client identifier (IP by default)
	IdentifierExtractor func(c echo.Context) (string, error)

	// Store implements storage for rate limit data
	Store Store

	// Limits per identifier
	RequestsPerUnit int
	Unit            time.Duration

	// Burst allows burst of requests up to this limit
	Burst int

	// KeyPrefix prefixes all keys - useful for namespacing
	KeyPrefix string

	// ErrorMessage is returned when rate limit exceeded
	ErrorMessage string

	// ErrorHandler handles rate limit errors
	ErrorHandler func(c echo.Context, e error) error

	// DenyHandler is called when rate limit is exceeded
	DenyHandler func(c echo.Context)
}

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	config Config
	store  Store
}

// Store defines the interface for rate limit storage
type Store interface {
	Allow(identifier string, limit int, window time.Duration) (bool, int, error)
}

// TokenBucketStore implements token bucket algorithm in memory
type TokenBucketStore struct {
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
	cleanupInterval time.Duration
}

type tokenBucket struct {
	tokens     int
	lastCheck time.Time
	mu        sync.Mutex
}

// NewTokenBucketStore creates a new in-memory token bucket store
func NewTokenBucketStore(cleanupInterval time.Duration) *TokenBucketStore {
	store := &TokenBucketStore{
		buckets:        make(map[string]*tokenBucket),
		cleanupInterval: cleanupInterval,
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Allow checks if request is allowed under rate limit
// Returns: allowed, remaining requests, error
func (s *TokenBucketStore) Allow(identifier string, limit int, window time.Duration) (bool, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	bucket, exists := s.buckets[identifier]

	if !exists {
		bucket = &tokenBucket{
			tokens:     limit - 1,
			lastCheck:  now,
		}
		s.buckets[identifier] = bucket
		return true, limit - 1, nil
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Calculate tokens to add based on elapsed time
	elapsed := now.Sub(bucket.lastCheck)
	tokensToAdd := int(elapsed / window) * limit

	if tokensToAdd > 0 {
		bucket.tokens = min(bucket.tokens+tokensToAdd, limit)
		bucket.lastCheck = now
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true, bucket.tokens, nil
	}

	return false, 0, nil
}

func (s *TokenBucketStore) cleanup() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		threshold := time.Now().Add(-s.cleanupInterval * 2)
		for key, bucket := range s.buckets {
			bucket.mu.Lock()
			if bucket.lastCheck.Before(threshold) {
				delete(s.buckets, key)
			}
			bucket.mu.Unlock()
		}
		s.mu.Unlock()
	}
}

// DefaultStore returns a store with default cleanup
func DefaultStore() Store {
	return NewTokenBucketStore(5 * time.Minute)
}

// NewRateLimiter creates a new rate limiter middleware
func NewRateLimiter(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	if config.Store == nil {
		config.Store = DefaultStore()
	}
	if config.ErrorMessage == "" {
		config.ErrorMessage = "rate limit exceeded"
	}

	_ = &RateLimiter{config: config}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			identifier := c.RealIP()
			if config.IdentifierExtractor != nil {
				var err error
				identifier, err = config.IdentifierExtractor(c)
				if err != nil {
					return next(c)
				}
			}

			limit := config.RequestsPerUnit
			if config.Burst > limit {
				limit = config.Burst
			}

			allowed, remaining, err := config.Store.Allow(
				identifier,
				limit,
				config.Unit,
			)

			// Set rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", itoa(limit))
			c.Response().Header().Set("X-RateLimit-Remaining", itoa(remaining))

			if err != nil {
				if config.ErrorHandler != nil {
					return config.ErrorHandler(c, err)
				}
			}

			if !allowed {
				c.Response().Header().Set("Retry-After", itoa(int(config.Unit.Seconds())))
				c.Response().Header().Set("X-RateLimit-Retry", "true")

				if config.DenyHandler != nil {
					config.DenyHandler(c)
					return nil
				}

				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"code":    429,
					"message": config.ErrorMessage,
				})
			}

			return next(c)
		}
	}
}

// RateLimiterFunc is a function that creates a rate limiter middleware
type RateLimiterFunc func(config Config) echo.MiddlewareFunc

// Global returns a global rate limiter for all requests
func Global(requestsPerUnit int, unit time.Duration) echo.MiddlewareFunc {
	return NewRateLimiter(Config{
		RequestsPerUnit: requestsPerUnit,
		Unit:            unit,
		Burst:           requestsPerUnit,
	})
}

// PerEndpoint returns rate limiter for specific endpoints
func PerEndpoint(requestsPerUnit int, unit time.Duration, burst int) echo.MiddlewareFunc {
	return NewRateLimiter(Config{
		RequestsPerUnit: requestsPerUnit,
		Unit:            unit,
		Burst:           burst,
	})
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	negative := false
	if i < 0 {
		negative = true
		i = -i
	}
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	if negative {
		result = "-" + result
	}
	return result
}
