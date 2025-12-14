package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gary-norman/forum/internal/models"
	"github.com/google/uuid"
)

// WithTracing adds request ID tracking and logs slow requests
func WithTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(human): Exercise 6 - Implement request tracing
		//
		// Context: Request tracing helps you debug issues by tracking individual requests
		// across your entire application. Every request gets a unique ID that flows through
		// logs, making it easy to find all operations related to a specific request.
		//
		// Your task: Implement request tracing with these steps:
		//
		// 1. Generate a unique request ID
		//    Hint: Use uuid.New().String() from the github.com/google/uuid package
		//
		// 2. Add request ID to context
		//    Hint: Use models.WithRequestID(r.Context(), requestID)
		//    This returns a new context with the request ID stored in it
		//
		// 3. Add request ID to response headers for debugging
		//    Hint: w.Header().Set("X-Request-ID", requestID)
		//    This lets clients reference the ID in bug reports
		//
		// 4. Track request duration (BONUS)
		//    - Record start time with time.Now()
		//    - Process the request with the new context: next.ServeHTTP(w, r.WithContext(ctx))
		//    - Calculate duration with time.Since(start)
		//    - If duration > 1 second, log a warning with request ID, method, path, and duration
		//
		// Pattern to follow:
		// requestID := uuid.New().String()
		// ctx := models.WithRequestID(r.Context(), requestID)
		// w.Header().Set("X-Request-ID", requestID)
		// start := time.Now()
		// next.ServeHTTP(w, r.WithContext(ctx))
		// duration := time.Since(start)
		// if duration > 1*time.Second { log.Printf("⚠️  SLOW REQUEST ...") }

		// Your implementation here:

	})
}
