package middleware

import (
	"context"
	"net/http"
	"time"
)

// WithTimeout creates a new HTTP handler that enforces a timeout on the request context.
func WithTimeout(next http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
