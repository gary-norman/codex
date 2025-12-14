package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker
type State int

const (
	// StateClosed means circuit is closed (normal operation)
	StateClosed State = iota
	// StateOpen means circuit is open (rejecting requests)
	StateOpen
	// StateHalfOpen means circuit is testing (allowing one request)
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "Closed"
	case StateOpen:
		return "Open"
	case StateHalfOpen:
		return "Half-Open"
	default:
		return "Unknown"
	}
}

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures      int           // Max failures before opening circuit
	resetTimeout     time.Duration // Time to wait before trying half-open
	windowSize       int           // Number of recent requests to track
	state            State
	failures         int
	lastFailureTime  time.Time
	halfOpenTestDone bool
	mu               sync.RWMutex
	// Sliding window for failure tracking
	recentRequests []bool // true = success, false = failure
	requestIndex   int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, windowSize int) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:    maxFailures,
		resetTimeout:   resetTimeout,
		windowSize:     windowSize,
		state:          StateClosed,
		recentRequests: make([]bool, 0, windowSize),
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if we can execute
	if !cb.canExecute() {
		return ErrCircuitOpen
	}

	// Execute the function
	err := fn()

	// Record result
	cb.recordResult(err == nil)

	return err
}

// canExecute checks if a request can proceed
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Always allow in closed state
		return true

	case StateOpen:
		// Check if enough time has passed to try half-open
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenTestDone = false
			return true
		}
		return false

	case StateHalfOpen:
		// In half-open, only allow one test request
		if !cb.halfOpenTestDone {
			cb.halfOpenTestDone = true
			return true
		}
		return false

	default:
		return false
	}
}

// recordResult records the result of a request
func (cb *CircuitBreaker) recordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Add to sliding window
	if len(cb.recentRequests) < cb.windowSize {
		cb.recentRequests = append(cb.recentRequests, success)
	} else {
		// Circular buffer: overwrite oldest entry
		cb.recentRequests[cb.requestIndex] = success
		cb.requestIndex = (cb.requestIndex + 1) % cb.windowSize
	}

	// Update failure count
	if !success {
		cb.failures++
		cb.lastFailureTime = time.Now()
	} else {
		// Success in half-open state closes the circuit
		if cb.state == StateHalfOpen {
			cb.state = StateClosed
			cb.failures = 0
			cb.recentRequests = make([]bool, 0, cb.windowSize)
			cb.requestIndex = 0
			return
		}
	}

	// Calculate failure rate in sliding window
	failureCount := 0
	for _, req := range cb.recentRequests {
		if !req {
			failureCount++
		}
	}

	// Open circuit if failure rate > 50%
	if len(cb.recentRequests) >= cb.windowSize {
		failureRate := float64(failureCount) / float64(len(cb.recentRequests))
		if failureRate > 0.5 {
			cb.state = StateOpen
			cb.lastFailureTime = time.Now()
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns current statistics
func (cb *CircuitBreaker) GetStats() (state State, failures int, failureRate float64) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	failureCount := 0
	for _, req := range cb.recentRequests {
		if !req {
			failureCount++
		}
	}

	rate := 0.0
	if len(cb.recentRequests) > 0 {
		rate = float64(failureCount) / float64(len(cb.recentRequests))
	}

	return cb.state, cb.failures, rate
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.recentRequests = make([]bool, 0, cb.windowSize)
	cb.requestIndex = 0
	cb.halfOpenTestDone = false
}
