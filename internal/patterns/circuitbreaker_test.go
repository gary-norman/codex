package patterns

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	// Circuit should start closed and allow requests
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected StateClosed, got %v", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	// Trigger 3 failures to open the circuit
	testErr := errors.New("test failure")
	for i := 0; i < 3; i++ {
		cb.Execute(func() error { return testErr })
	}

	if cb.State() != StateOpen {
		t.Errorf("Expected StateOpen after 3 failures, got %v", cb.State())
	}
	if cb.Failures() != 3 {
		t.Errorf("Expected 3 failures, got %v", cb.Failures())
	}
}

func TestCircuitBreaker_BlocksWhenOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 200*time.Millisecond)

	// Open the circuit
	testErr := errors.New("test failure")
	cb.Execute(func() error { return testErr })
	cb.Execute(func() error { return testErr })

	// Next request should be blocked
	err := cb.Execute(func() error { return nil })
	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	// Open the circuit
	testErr := errors.New("test failure")
	cb.Execute(func() error { return testErr })
	cb.Execute(func() error { return testErr })

	if cb.State() != StateOpen {
		t.Fatalf("Circuit should be open")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next request should transition to half-open
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Errorf("Expected nil after timeout, got %v", err)
	}

	// Success should close the circuit
	if cb.State() != StateClosed {
		t.Errorf("Expected StateClosed after successful half-open request, got %v", cb.State())
	}
	if cb.Failures() != 0 {
		t.Errorf("Expected failures reset to 0, got %v", cb.Failures())
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	// Open the circuit
	testErr := errors.New("test failure")
	cb.Execute(func() error { return testErr })
	cb.Execute(func() error { return testErr })

	// Wait for timeout to enter half-open
	time.Sleep(60 * time.Millisecond)

	// Half-open request fails
	cb.Execute(func() error { return testErr })

	// Circuit should reopen
	if cb.State() != StateOpen {
		t.Errorf("Expected StateOpen after half-open failure, got %v", cb.State())
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker(10, 100*time.Millisecond)

	// Concurrent requests should not cause race conditions
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			cb.Execute(func() error { return nil })
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected StateClosed, got %v", cb.State())
	}
}
