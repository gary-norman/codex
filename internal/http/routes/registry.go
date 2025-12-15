package routes

import (
	"context"
	"github.com/gary-norman/forum/internal/app"
	h "github.com/gary-norman/forum/internal/http/handlers"
	"github.com/gary-norman/forum/internal/http/websocket"
)

type RouteHandler struct {
	App       *app.App
	Auth      *h.AuthHandler
	Channel   *h.ChannelHandler
	Comment   *h.CommentHandler
	Home      *h.HomeHandler
	Post      *h.PostHandler
	Reaction  *h.ReactionHandler
	Search    *h.SearchHandler
	Session   *h.SessionHandler
	User      *h.UserHandler
	Mod       *h.ModHandler
	Websocket *websocket.Manager
}

func NewCommentHandler(app *app.App, reaction *h.ReactionHandler) *h.CommentHandler {
	return &h.CommentHandler{
		App:      app,
		Reaction: reaction,
	}
}

func NewWebsocketHandler(
	app *app.App,
	user *h.UserHandler,
	post *h.PostHandler,
	comment *h.CommentHandler,
	reaction *h.ReactionHandler,
	channel *h.ChannelHandler,
	mod *h.ModHandler,
	ctx context.Context,
) *websocket.Manager {
	ws := websocket.NewManager(ctx)

	ws.App = app
	ws.User = user
	ws.Post = post
	ws.Comment = comment
	ws.Reaction = reaction
	ws.Channel = channel
	ws.Mod = mod

	return ws
}

func NewReactionHandler(app *app.App) *h.ReactionHandler {
	return &h.ReactionHandler{
		App: app,
	}
}

func NewChannelHandler(app *app.App, comment *h.CommentHandler, reaction *h.ReactionHandler) *h.ChannelHandler {
	return &h.ChannelHandler{
		App:      app,
		Comment:  comment,
		Reaction: reaction,
	}
}

func NewUserHandler(app *app.App, channel *h.ChannelHandler, comment *h.CommentHandler, reaction *h.ReactionHandler) *h.UserHandler {
	return &h.UserHandler{
		App:      app,
		Comment:  comment,
		Channel:  channel,
		Reaction: reaction,
	}
}

func NewPostHandler(app *app.App, channel *h.ChannelHandler, comment *h.CommentHandler, reaction *h.ReactionHandler) *h.PostHandler {
	return &h.PostHandler{
		App:      app,
		Channel:  channel,
		Comment:  comment,
		Reaction: reaction,
	}
}

func NewHomeHandler(app *app.App, channel *h.ChannelHandler, comment *h.CommentHandler, post *h.PostHandler, reaction *h.ReactionHandler) *h.HomeHandler {
	return &h.HomeHandler{
		App:      app,
		Channel:  channel,
		Comment:  comment,
		Post:     post,
		Reaction: reaction,
	}
}

func NewSessionHandler(app *app.App) *h.SessionHandler {
	return &h.SessionHandler{
		App: app,
	}
}

func NewAuthHandler(app *app.App, session *h.SessionHandler) *h.AuthHandler {
	return &h.AuthHandler{
		App:     app,
		Session: session,
	}
}

func NewSearchHandler(app *app.App) *h.SearchHandler {
	return &h.SearchHandler{
		App: app,
	}
}

func NewModHandler(app *app.App, channel *h.ChannelHandler, user *h.UserHandler) *h.ModHandler {
	return &h.ModHandler{
		App:     app,
		Channel: channel,
		User:    user,
	}
}

func NewRouteHandler(app *app.App) *RouteHandler {
	// Step 1: Create top-level (flat) handlers without nested deps first
	sessionHandler := NewSessionHandler(app)
	reactionHandler := NewReactionHandler(app)
	authHandler := NewAuthHandler(app, sessionHandler)

	// Step 2: Create nested handlers with their deps injected
	commentHandler := NewCommentHandler(app, reactionHandler)
	channelHandler := NewChannelHandler(app, commentHandler, reactionHandler)
	userHandler := NewUserHandler(app, channelHandler, commentHandler, reactionHandler)
	postHandler := NewPostHandler(app, channelHandler, commentHandler, reactionHandler)
	homeHandler := NewHomeHandler(app, channelHandler, commentHandler, postHandler, reactionHandler)
	modHandler := NewModHandler(app, channelHandler, userHandler)
	websocketHandler := NewWebsocketHandler(app, userHandler, postHandler, commentHandler, reactionHandler, channelHandler, modHandler, context.Background())
	searchHandler := NewSearchHandler(app)

	// Step 3: Return fully wired router
	return &RouteHandler{
		App:       app,
		Auth:      authHandler,
		Channel:   channelHandler,
		Comment:   commentHandler,
		Home:      homeHandler,
		Post:      postHandler,
		Reaction:  reactionHandler,
		Search:    searchHandler,
		Session:   sessionHandler,
		User:      userHandler,
		Mod:       modHandler,
		Websocket: websocketHandler,
	}
}
