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
// Message is neutral, args are blue
func LogInfo(msg string, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := formatMessageWithBlueArgs(msg, args...)
	formatted := fmt.Sprintf("%s%s%s ℹ️  %s",
		Colors.Subtext0, timestamp, Colors.Reset,
		formattedMsg)
	log.Println(formatted)
}

// LogWarn logs a warning message with timestamp, icon, and color
// Message is neutral, args are blue
func LogWarn(msg string, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := formatMessageWithBlueArgs(msg, args...)
	formatted := fmt.Sprintf("%s%s%s ⚠️  %s",
		Colors.Subtext0, timestamp, Colors.Reset,
		formattedMsg)
	log.Println(formatted)
}

// LogError logs an error message with timestamp, icon, color, and optional error
// Message is neutral, args are blue, error is neutral
func LogError(msg string, err error, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := formatMessageWithBlueArgs(msg, args...)
	formatted := fmt.Sprintf("%s%s%s ❌ %s",
		Colors.Subtext0, timestamp, Colors.Reset,
		formattedMsg)
	if err != nil {
		formatted += fmt.Sprintf(" %s%v%s", Colors.Text, err, Colors.Reset)
	}
	log.Println(formatted)
}

// formatMessageWithBlueArgs formats a message with neutral text and blue arguments
func formatMessageWithBlueArgs(msg string, args ...any) string {
	// Color the args blue
	coloredArgs := make([]any, len(args))
	for i, arg := range args {
		coloredArgs[i] = fmt.Sprintf("%s%v%s", Colors.Blue, arg, Colors.Reset)
	}

	// Format the message with colored args, then wrap in neutral text color
	return fmt.Sprintf("%s%s%s", Colors.Text, fmt.Sprintf(msg, coloredArgs...), Colors.Reset)
}

// LogInfoWithContext logs an info message with request ID from context
// Request ID is colored green, message is neutral, args are blue
func LogInfoWithContext(ctx context.Context, msg string, args ...any) {
	requestID := GetRequestID(ctx)
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := formatMessageWithBlueArgs(msg, args...)

	if requestID != "" {
		formatted := fmt.Sprintf("%s%s%s ℹ️  %s[%s]%s %s",
			Colors.Subtext0, timestamp, Colors.Reset,
			Colors.Green, requestID, Colors.Reset,
			formattedMsg)
		log.Println(formatted)
	} else {
		LogInfo(msg, args...)
	}
}

// LogWarnWithContext logs a warning message with request ID from context
// Request ID is colored orange, message is neutral, args are blue
func LogWarnWithContext(ctx context.Context, msg string, args ...any) {
	requestID := GetRequestID(ctx)
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := formatMessageWithBlueArgs(msg, args...)

	if requestID != "" {
		formatted := fmt.Sprintf("%s%s%s ⚠️  %s[%s]%s %s",
			Colors.Subtext0, timestamp, Colors.Reset,
			Colors.Peach, requestID, Colors.Reset,
			formattedMsg)
		log.Println(formatted)
	} else {
		LogWarn(msg, args...)
	}
}

// LogErrorWithContext logs an error message with request ID from context
// Request ID is colored red, message is neutral, args are blue, error is neutral
func LogErrorWithContext(ctx context.Context, msg string, err error, args ...any) {
	requestID := GetRequestID(ctx)
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := formatMessageWithBlueArgs(msg, args...)

	if requestID != "" {
		formatted := fmt.Sprintf("%s%s%s ❌ %s[%s]%s %s",
			Colors.Subtext0, timestamp, Colors.Reset,
			Colors.Red, requestID, Colors.Reset,
			formattedMsg)
		if err != nil {
			formatted += fmt.Sprintf(" %s%v%s", Colors.Text, err, Colors.Reset)
		}
		log.Println(formatted)
	} else {
		LogError(msg, err, args...)
	}
}
