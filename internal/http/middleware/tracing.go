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
		// Generate unique request ID
		requestID := uuid.New().String()

		// Add request ID to context
		ctx := models.WithRequestID(r.Context(), requestID)

		// Add request ID to response headers for debugging
		w.Header().Set("X-Request-ID", requestID)

		// Track request duration (for slow request logging)
		start := time.Now()

		// Process request with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))

		// Calculate duration
		duration := time.Since(start)

		// BONUS: Log slow requests (>1 second)
		if duration > 1*time.Second {
			log.Printf("⚠️  SLOW REQUEST [%s] %s %s - took %v",
				requestID, r.Method, r.URL.Path, duration)
		}
	})
}
