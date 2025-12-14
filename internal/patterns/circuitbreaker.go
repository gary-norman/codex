package patterns

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker's current state
type State int

const (
	StateClosed   State = iota // Normal operation, requests pass through
	StateOpen                  // Failure threshold exceeded, requests blocked
	StateHalfOpen              // Testing if service recovered
)

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// CircuitBreaker prevents cascading failures by tracking errors and blocking requests
type CircuitBreaker struct {
	maxFailures  uint32        // Failures before opening circuit
	timeout      time.Duration // How long to wait before testing recovery
	state        State
	failures     uint32
	lastFailTime time.Time
	mu           sync.RWMutex
}

// NewCircuitBreaker creates a circuit breaker with specified thresholds
func NewCircuitBreaker(maxFailures uint32, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       StateClosed,
	}
}

// Execute runs a function through the circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if circuit allows the request
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function and track result
	err := fn()
	cb.afterRequest(err)
	return err
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.failures = 0
			return nil
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		return nil
	}
	return nil
}

// afterRequest updates circuit breaker state based on request result
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		// Request failed
		cb.failures++
		cb.lastFailTime = time.Now()

		// If in half-open state and request fails, immediately reopen
		if cb.state == StateHalfOpen {
			cb.state = StateOpen
		} else if cb.failures >= cb.maxFailures {
			// Trip the circuit if failure threshold exceeded
			cb.state = StateOpen
		}
	} else {
		// Request succeeded
		if cb.state == StateHalfOpen {
			// Recovery confirmed, close the circuit
			cb.state = StateClosed
		}
		cb.failures = 0
	}
}

// State returns current circuit breaker state (for monitoring/testing)
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Failures returns current failure count (for monitoring/testing)
func (cb *CircuitBreaker) Failures() uint32 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}
