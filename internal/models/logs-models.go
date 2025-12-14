package models

import (
	"context"
	"fmt"
	"log"
	"time"
)

// LogContextKey is the type used for context keys
type LogContextKey string

const RequestIDKey LogContextKey = "requestID"

// WithRequestID is a helper function that adds a Request ID to a context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// LogInfo logs an info message with timestamp, icon, and color
func LogInfo(msg string, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	formatted := fmt.Sprintf("%s%s%s ℹ️  %s%s%s",
		Colors.Subtext0, timestamp, Colors.Reset,
		Colors.Green, fmt.Sprintf(msg, args...), Colors.Reset)
	log.Println(formatted)
}

// LogWarn logs a warning message with timestamp, icon, and color
func LogWarn(msg string, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	formatted := fmt.Sprintf("%s%s%s ⚠️  %s%s%s",
		Colors.Subtext0, timestamp, Colors.Reset,
		Colors.Peach, fmt.Sprintf(msg, args...), Colors.Reset)
	log.Println(formatted)
}

// LogError logs an error message with timestamp, icon, color, and optional error
func LogError(msg string, err error, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	formatted := fmt.Sprintf("%s%s%s ❌ %s%s%s",
		Colors.Subtext0, timestamp, Colors.Reset,
		Colors.Red, fmt.Sprintf(msg, args...), Colors.Reset)
	if err != nil {
		formatted += fmt.Sprintf(" %s%v%s", Colors.Text, err, Colors.Reset)
	}
	log.Println(formatted)
}

// LogInfoWithContext logs an info message with request ID from context
func LogInfoWithContext(ctx context.Context, msg string, args ...any) {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		msg = fmt.Sprintf("[%s] %s", requestID, msg)
	}
	LogInfo(msg, args...)
}

// LogWarnWithContext logs a warning message with request ID from context
func LogWarnWithContext(ctx context.Context, msg string, args ...any) {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		msg = fmt.Sprintf("[%s] %s", requestID, msg)
	}
	LogWarn(msg, args...)
}

// LogErrorWithContext logs an error message with request ID from context
func LogErrorWithContext(ctx context.Context, msg string, err error, args ...any) {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		msg = fmt.Sprintf("[%s] %s", requestID, msg)
	}
	LogError(msg, err, args...)
}
