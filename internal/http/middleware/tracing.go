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
		requestID := uuid.New().String()
		ctx := models.WithRequestID(r.Context(), requestID)
		w.Header().Set("X-Request-ID", requestID)
		start := time.Now()
		next.ServeHTTP(w, r.WithContext(ctx))
		duration := time.Since(start)
		if duration > 1*time.Second {
			log.Printf("⚠️  SLOW REQUEST [%s] %s - took %v", requestID, r.URL.Path, duration)
		}
	})
}
