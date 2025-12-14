// Package routes provides the HTTP routing for the application.
package routes

import (
	"net/http"
	"time"

	"github.com/gary-norman/forum/internal/app"
	// "github.com/gary-norman/forum/internal/http/handlers"
	mw "github.com/gary-norman/forum/internal/http/middleware"
	"github.com/gary-norman/forum/internal/workers"
)

func NewRouter(app *app.App, loggerPool *workers.LoggerPool) http.Handler {
	mux := http.NewServeMux()
	r := NewRouteHandler(app)

	// Static
	// handlers.MuxHandler(mux, "assets")
	// handlers.MuxHandler(mux, "db")
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	mux.Handle("/db/", http.StripPrefix("/db/", http.FileServer(http.Dir("./db"))))

	// Core routes
	mux.HandleFunc("POST /register", r.Auth.Register)
	mux.HandleFunc("POST /login", r.Auth.Login)
	mux.HandleFunc("POST /logout", r.Auth.Logout)
	mux.HandleFunc("POST /protected", r.Auth.Protected)
	mux.Handle("/", mw.WithUser(http.HandlerFunc(r.Home.RenderIndex), r.App))
	mux.Handle("/home", mw.WithUser(http.HandlerFunc(r.Home.GetHome), r.App))
	mux.Handle("/{invalidString}", mw.WithUser(http.HandlerFunc(r.Home.RenderIndex), r.App))
	// mux.HandleFunc("GET /posts/create", r.Post.CreatePost)
	mux.Handle("GET /search", mw.WithUser(http.HandlerFunc(r.Search.Search), r.App))
	mux.Handle("GET /post/{postId}", mw.WithUser(http.HandlerFunc(r.Post.GetThisPost), r.App))
	mux.Handle("GET /user/{userId}", mw.WithUser(http.HandlerFunc(r.User.GetThisUser), r.App))
	mux.Handle("GET /channel/{channelId}", mw.WithUser(http.HandlerFunc(r.Channel.GetThisChannel), r.App))
	// mux.Handle("GET /comments/{commentId}", mw.WithUser(http.HandlerFunc(r.Comment.GetThisComment), r.App))
	mux.Handle("POST /posts/create", mw.WithUser(http.HandlerFunc(r.Post.StorePost), r.App))
	mux.Handle("POST /channels/create", mw.WithUser(http.HandlerFunc(r.Channel.StoreChannel), r.App))
	mux.Handle("POST /store-reaction", mw.WithUser(http.HandlerFunc(r.Reaction.StoreReaction), r.App))
	mux.Handle("POST /edituser", mw.WithUser(http.HandlerFunc(r.User.EditUserDetails), r.App))
	mux.Handle("POST /channels/join", mw.WithUser(http.HandlerFunc(r.Channel.StoreMembership), r.App))
	mux.Handle("POST /channels/add-rules/{channelId}", mw.WithUser(http.HandlerFunc(r.Channel.CreateAndInsertRule), r.App))
	mux.Handle("POST /cdx/post/{postId}/store-comment", mw.WithUser(http.HandlerFunc(r.Comment.StoreComment), r.App))

	// Apply middleware chain: Tracing (outermost) -> Logging -> Timeout
	// Order matters! Tracing must be first so request ID exists before logging
	timeoutHandler := mw.WithTimeout(mux, 10*time.Second)
	loggingHandler := mw.LoggingEnhanced(loggerPool)(timeoutHandler)
	return mw.WithTracing(loggingHandler)
}
