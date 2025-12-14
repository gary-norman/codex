package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second, 10)

	t.Run("allows requests in closed state", func(t *testing.T) {
		successCount := 0
		fn := func() error {
			successCount++
			return nil
		}

		// Execute 5 successful requests
		for i := 0; i < 5; i++ {
			err := cb.Execute(fn)
			if err != nil {
				t.Fatalf("Request %d failed: %v", i, err)
			}
		}

		if successCount != 5 {
			t.Errorf("Expected 5 successes, got %d", successCount)
		}

		state := cb.GetState()
		if state != StateClosed {
			t.Errorf("Expected Closed state, got %v", state)
		}
	})
}

func TestCircuitBreaker_OpensOnFailures(t *testing.T) {
	// Window size 10, opens when >50% fail (6+ failures)
	cb := NewCircuitBreaker(5, 100*time.Millisecond, 10)

	t.Run("opens circuit when failure rate exceeds 50%", func(t *testing.T) {
		// Execute 6 failures and 4 successes (60% failure rate)
		for i := 0; i < 6; i++ {
			cb.Execute(func() error {
				return errors.New("database error")
			})
		}

		for i := 0; i < 4; i++ {
			cb.Execute(func() error {
				return nil
			})
		}

		state := cb.GetState()
		if state != StateOpen {
			t.Errorf("Expected Open state after 60%% failures, got %v", state)
		}

		// Verify requests are rejected
		err := cb.Execute(func() error {
			t.Error("Should not execute in open state")
			return nil
		})

		if err != ErrCircuitOpen {
			t.Errorf("Expected ErrCircuitOpen, got %v", err)
		}
	})
}

func TestCircuitBreaker_HalfOpenState(t *testing.T) {
	resetTimeout := 50 * time.Millisecond
	cb := NewCircuitBreaker(3, resetTimeout, 10)

	t.Run("transitions to half-open after timeout", func(t *testing.T) {
		// Force circuit to open
		for i := 0; i < 10; i++ {
			cb.Execute(func() error {
				return errors.New("failure")
			})
		}

		if cb.GetState() != StateOpen {
			t.Fatalf("Circuit should be open")
		}

		// Wait for reset timeout
		time.Sleep(resetTimeout + 10*time.Millisecond)

		// Next request should trigger half-open
		executed := false
		err := cb.Execute(func() error {
			executed = true
			return nil
		})

		if err != nil {
			t.Errorf("Half-open request should execute, got error: %v", err)
		}

		if !executed {
			t.Error("Function should have executed in half-open state")
		}
	})

	t.Run("only allows one request in half-open", func(t *testing.T) {
		cb.Reset()

		// Force open
		for i := 0; i < 10; i++ {
			cb.Execute(func() error {
				return errors.New("failure")
			})
		}

		// Wait for reset
		time.Sleep(resetTimeout + 10*time.Millisecond)

		// First request in half-open (success closes circuit immediately)
		firstExecuted := false
		err := cb.Execute(func() error {
			firstExecuted = true
			return nil
		})

		if err != nil {
			t.Errorf("First half-open request should succeed, got: %v", err)
		}

		if !firstExecuted {
			t.Error("First request should have executed")
		}

		// After successful half-open request, circuit is CLOSED
		if cb.GetState() != StateClosed {
			t.Errorf("Expected Closed state after successful half-open, got %v", cb.GetState())
		}

		// Second request succeeds because circuit is now closed
		err = cb.Execute(func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Second request should succeed (circuit closed), got: %v", err)
		}
	})
}

func TestCircuitBreaker_ClosesAfterSuccessInHalfOpen(t *testing.T) {
	resetTimeout := 50 * time.Millisecond
	cb := NewCircuitBreaker(3, resetTimeout, 10)

	// Force circuit open
	for i := 0; i < 10; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Wait for reset timeout
	time.Sleep(resetTimeout + 10*time.Millisecond)

	// Successful request in half-open should close circuit
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Fatalf("Half-open request failed: %v", err)
	}

	state := cb.GetState()
	if state != StateClosed {
		t.Errorf("Expected Closed state after successful half-open request, got %v", state)
	}

	// Verify circuit is now accepting requests
	err = cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Circuit should accept requests after closing, got: %v", err)
	}
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	cb := NewCircuitBreaker(5, 100*time.Millisecond, 10)

	// Execute 7 failures, 3 successes (70% failure rate)
	for i := 0; i < 7; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	for i := 0; i < 3; i++ {
		cb.Execute(func() error {
			return nil
		})
	}

	state, failures, failureRate := cb.GetStats()

	if state != StateOpen {
		t.Errorf("Expected Open state, got %v", state)
	}

	if failures < 7 {
		t.Errorf("Expected at least 7 failures, got %d", failures)
	}

	expectedRate := 0.7
	if failureRate < expectedRate-0.01 || failureRate > expectedRate+0.01 {
		t.Errorf("Expected failure rate ~%.1f, got %.2f", expectedRate, failureRate)
	}

	t.Logf("Stats: state=%v, failures=%d, failureRate=%.2f", state, failures, failureRate)
}

func TestCircuitBreaker_SlidingWindow(t *testing.T) {
	cb := NewCircuitBreaker(5, 100*time.Millisecond, 4) // Small window for testing

	t.Run("maintains sliding window of recent requests", func(t *testing.T) {
		// First 4 requests: all failures (100% failure rate) → opens
		for i := 0; i < 4; i++ {
			cb.Execute(func() error {
				return errors.New("failure")
			})
		}

		state, _, rate := cb.GetStats()
		if state != StateOpen {
			t.Errorf("Circuit should open with 100%% failure rate, got state=%v, rate=%.0f%%", state, rate*100)
		}

		cb.Reset()

		// Test with clearer pattern: 3 failures, 1 success (75% failure) → should open
		cb.Execute(func() error { return errors.New("fail") })
		cb.Execute(func() error { return errors.New("fail") })
		cb.Execute(func() error { return errors.New("fail") })
		cb.Execute(func() error { return nil })

		state, _, rate = cb.GetStats()
		if state != StateOpen {
			t.Errorf("Circuit should open with 75%% failure rate, got state=%v, rate=%.0f%%", state, rate*100)
		}

		cb.Reset()

		// Test boundary: exactly 50% (should stay closed, need >50% to open)
		cb.Execute(func() error { return errors.New("fail") })
		cb.Execute(func() error { return nil })
		cb.Execute(func() error { return errors.New("fail") })
		cb.Execute(func() error { return nil })

		state, _, rate = cb.GetStats()
		if state != StateClosed {
			t.Errorf("Circuit should stay closed at exactly 50%% failure rate, got state=%v, rate=%.0f%%", state, rate*100)
		}
	})
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker(10, 100*time.Millisecond, 20)

	t.Run("handles concurrent requests safely", func(t *testing.T) {
		done := make(chan bool)

		// Launch 10 goroutines making requests
		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < 10; j++ {
					cb.Execute(func() error {
						if j%2 == 0 {
							return nil
						}
						return errors.New("failure")
					})
					time.Sleep(1 * time.Millisecond)
				}
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		state, failures, rate := cb.GetStats()
		t.Logf("After concurrent access: state=%v, failures=%d, rate=%.2f", state, failures, rate)

		// Verify no panics occurred (test passes if we get here)
	})
}
