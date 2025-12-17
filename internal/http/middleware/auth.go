package middleware

import (
	"context"
	"net/http"

	"github.com/gary-norman/forum/internal/app"
	"github.com/gary-norman/forum/internal/colors"
	"github.com/gary-norman/forum/internal/models"
)

var (
	authColors, _ = colors.UseFlavor("Mocha")
)

// Create a custom key type to avoid conflicts in context
type contextKey string

const userContextKey = contextKey("currentUser")

// Middleware to add the user to the request context
func WithUser(next http.Handler, app *app.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var currentUser *models.User
		cookie, err := r.Cookie("username")
		if err != nil || cookie.Value == "" {
			models.LogWarn("No session cookie found in request: %v", err)
			next.ServeHTTP(w, r)
			return
		}
		user, err := app.Users.GetUserByUsername(ctx, cookie.Value, "WithUser")
		if err != nil {
			models.LogWarn("No user found for username %s: %v", cookie.Value, err)
			next.ServeHTTP(w, r)
			return
		}
		currentUser = user

		// Store user in context
		ctx = context.WithValue(r.Context(), userContextKey, currentUser)
		// Pass modified request with context to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
