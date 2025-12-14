package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gary-norman/forum/internal/models"
)

func TestWithTracing(t *testing.T) {
	t.Run("adds request ID to context", func(t *testing.T) {
		// Create test handler that checks for request ID in context
		var capturedRequestID string
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID, ok := r.Context().Value(models.RequestIDKey).(string)
			if !ok || requestID == "" {
				t.Error("Request ID not found in context")
			}
			capturedRequestID = requestID
			w.WriteHeader(http.StatusOK)
		})

		// Wrap with tracing middleware
		tracedHandler := WithTracing(handler)

		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// Execute request
		tracedHandler.ServeHTTP(w, req)

		// Verify request ID was added to context
		if capturedRequestID == "" {
			t.Error("Request ID was not added to context")
		}
	})

	t.Run("adds X-Request-ID header to response", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		tracedHandler := WithTracing(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		tracedHandler.ServeHTTP(w, req)

		// Verify X-Request-ID header was added
		requestID := w.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("X-Request-ID header not set in response")
		}

		// Verify it's a valid UUID format
		if len(requestID) != 36 {
			t.Errorf("Request ID has unexpected length: got %d, want 36", len(requestID))
		}
	})

	t.Run("logs slow requests", func(t *testing.T) {
		// Create slow handler (>1 second)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

		tracedHandler := WithTracing(handler)

		req := httptest.NewRequest(http.MethodGet, "/slow-endpoint", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		tracedHandler.ServeHTTP(w, req)
		duration := time.Since(start)

		// Verify request actually took >1 second
		if duration < 1*time.Second {
			t.Errorf("Test handler didn't run long enough: %v", duration)
		}

		// Note: We can't easily capture log output in this test
		// In production, you'd see: "⚠️  SLOW REQUEST [uuid] GET /slow-endpoint - took 1.1s"
		t.Logf("Slow request completed in %v (should have been logged)", duration)
	})

	t.Run("request ID is unique per request", func(t *testing.T) {
		var requestID1, requestID2 string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		tracedHandler := WithTracing(handler)

		// First request
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		w1 := httptest.NewRecorder()
		tracedHandler.ServeHTTP(w1, req1)
		requestID1 = w1.Header().Get("X-Request-ID")

		// Second request
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		w2 := httptest.NewRecorder()
		tracedHandler.ServeHTTP(w2, req2)
		requestID2 = w2.Header().Get("X-Request-ID")

		// Verify request IDs are different
		if requestID1 == requestID2 {
			t.Error("Request IDs should be unique for each request")
		}
	})
}

func TestTracingWithLoggingEnhanced(t *testing.T) {
	t.Run("request ID flows through middleware chain", func(t *testing.T) {
		var capturedRequestID string

		// Inner handler captures request ID from context
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID, _ := r.Context().Value(models.RequestIDKey).(string)
			capturedRequestID = requestID
			w.WriteHeader(http.StatusOK)
		})

		// Apply tracing middleware
		tracedHandler := WithTracing(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		tracedHandler.ServeHTTP(w, req)

		// Verify request ID from context matches header
		headerRequestID := w.Header().Get("X-Request-ID")
		if capturedRequestID != headerRequestID {
			t.Errorf("Context request ID (%s) doesn't match header (%s)", capturedRequestID, headerRequestID)
		}
	})
}
