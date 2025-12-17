package middleware

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/workers"
)

// ResponseWriter wrapper to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// Hijack implements http.Hijacker interface for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}

// LoggingEnhanced is a middleware that logs detailed request metrics to the database
func LoggingEnhanced(loggerPool *workers.LoggerPool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start timing
			start := time.Now()

			// Wrap response writer to capture status code and bytes
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     200, // Default status code
			}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start).Milliseconds()

			user, ok := r.Context().Value("user").(*models.User)
			var userID models.UUIDField
			if ok && user != nil {
				userID = user.ID
			} else {
				userID = models.ZeroUUIDField() // Nil UUID for anonymous
			}

			// Build request log entry
			requestLog := models.RequestLog{
				Timestamp:  start,
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: wrapped.statusCode,
				Duration:   duration,
				UserID:     userID,
				IPAddress:  getClientIP(r),
				UserAgent:  r.UserAgent(),
				Referer:    r.Referer(),
				BytesSent:  wrapped.bytesWritten,
			}

			// Submit log asynchronously (non-blocking)
			if err := loggerPool.LogRequest(requestLog); err != nil {
				// If queue is full, just log to console - don't slow down the request!
				log.Printf("Warning: Failed to queue request log: %v\n", err)
			}

			// Also log to console for immediate visibility (optional)
			log.Printf("[%s] %s - %d (%dms)\n",
				r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// getClientIP extracts the real client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
