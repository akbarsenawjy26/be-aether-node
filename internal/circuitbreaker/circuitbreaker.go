package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTooManyRequests is returned when too many requests are being processed
	ErrTooManyRequests = errors.New("too many requests")
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// Config holds circuit breaker configuration
type Config struct {
	Name        string

	// Failure threshold to trip the circuit
	FailureThreshold int

	// Success threshold to close the circuit (in half-open state)
	SuccessThreshold int

	// Timeout duration before transitioning from open to half-open
	Timeout time.Duration

	// Maximum concurrent requests in half-open state
	MaxConcurrentRequests int

	// OnStateChange is called when state changes
	OnStateChange func(name string, from, to State)
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	name   string
	config Config

	mu     sync.RWMutex
	state  State

	failures    int
	successes   int
	lastFailure time.Time

	semaphore chan struct{} // limits concurrent requests in half-open
}

// New creates a new circuit breaker
func New(config Config) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:   config.Name,
		config: config,
		state:  StateClosed,
	}

	if cb.config.FailureThreshold == 0 {
		cb.config.FailureThreshold = 5
	}
	if cb.config.SuccessThreshold == 0 {
		cb.config.SuccessThreshold = 2
	}
	if cb.config.Timeout == 0 {
		cb.config.Timeout = 30 * time.Second
	}
	if cb.config.MaxConcurrentRequests == 0 {
		cb.config.MaxConcurrentRequests = 1
	}

	cb.semaphore = make(chan struct{}, cb.config.MaxConcurrentRequests)

	return cb
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Allow checks if request is allowed through the circuit breaker
func (cb *CircuitBreaker) Allow() error {
	state := cb.State()

	switch state {
	case StateClosed:
		return nil

	case StateOpen:
		// Check if timeout has passed
		cb.mu.RLock()
		timeout := cb.config.Timeout
		lastFailure := cb.lastFailure
		cb.mu.RUnlock()

		if time.Since(lastFailure) >= timeout {
			cb.transitionTo(StateHalfOpen)
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// Try to acquire semaphore
		select {
		case cb.semaphore <- struct{}{}:
			return nil
		default:
			return ErrTooManyRequests
		}
	}

	return nil
}

// Release releases a slot in half-open state
func (cb *CircuitBreaker) Release() {
	<-cb.semaphore
}

// RecordSuccess records a successful call
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.transitionToLocked(StateClosed)
		}
	} else if cb.state == StateClosed {
		// Reset failures on success
		cb.failures = 0
	}
}

// RecordFailure records a failed call
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailure = time.Now()

	if cb.state == StateHalfOpen {
		cb.transitionToLocked(StateOpen)
	} else if cb.state == StateClosed {
		cb.failures++
		if cb.failures >= cb.config.FailureThreshold {
			cb.transitionToLocked(StateOpen)
		}
	}
}

// Execute runs the function through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := cb.Allow(); err != nil {
		return err
	}

	defer cb.Release()

	err := fn(ctx)
	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// transitionTo changes state with read lock
func (cb *CircuitBreaker) transitionTo(newState State) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transitionToLocked(newState)
}

// transitionToLocked changes state - caller must hold lock
func (cb *CircuitBreaker) transitionToLocked(newState State) {
	oldState := cb.state

	if oldState == newState {
		return
	}

	cb.state = newState

	// Reset counters on state transition
	cb.failures = 0
	cb.successes = 0

	// Clear semaphore on open
	if newState == StateOpen {
		// Drain the semaphore
		for {
			select {
			case <-cb.semaphore:
			default:
				return
			}
		}
	}

	// Clear semaphore on close
	if newState == StateClosed {
		// Drain the semaphore
		for {
			select {
			case <-cb.semaphore:
			default:
				return
			}
		}
	}

	if cb.config.OnStateChange != nil {
		cb.config.OnStateChange(cb.name, oldState, newState)
	}
}

// CircuitBreakerGroup manages multiple circuit breakers
type CircuitBreakerGroup struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
	defaultConfig Config
}

// NewGroup creates a new circuit breaker group
func NewGroup(config Config) *CircuitBreakerGroup {
	return &CircuitBreakerGroup{
		breakers:     make(map[string]*CircuitBreaker),
		defaultConfig: config,
	}
}

// Get returns or creates a circuit breaker for the given name
func (g *CircuitBreakerGroup) Get(name string) *CircuitBreaker {
	g.mu.RLock()
	cb, exists := g.breakers[name]
	g.mu.RUnlock()

	if exists {
		return cb
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists = g.breakers[name]; exists {
		return cb
	}

	config := g.defaultConfig
	config.Name = name
	cb = New(config)
	g.breakers[name] = cb

	return cb
}

// Stats returns stats for all circuit breakers
func (g *CircuitBreakerGroup) Stats() map[string]Stats {
	g.mu.RLock()
	defer g.mu.RUnlock()

	stats := make(map[string]Stats)
	for name, cb := range g.breakers {
		stats[name] = cb.Stats()
	}
	return stats
}

// Stats holds circuit breaker statistics
type Stats struct {
	Name        string
	State       string
	Failures    int
	Successes   int
	LastFailure time.Time
}

// Stats returns the current stats of the circuit breaker
func (cb *CircuitBreaker) Stats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return Stats{
		Name:        cb.name,
		State:       cb.state.String(),
		Failures:    cb.failures,
		Successes:   cb.successes,
		LastFailure: cb.lastFailure,
	}
}
