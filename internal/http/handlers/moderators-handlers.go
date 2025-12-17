package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gary-norman/forum/internal/app"
	mw "github.com/gary-norman/forum/internal/http/middleware"
	"github.com/gary-norman/forum/internal/models"
)

type ModHandler struct {
	App     *app.App
	Channel *ChannelHandler
	User    *UserHandler
}
type APIResponse struct {
	StatusCode int
	Message    string
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, message string) {
	resp := APIResponse{
		StatusCode: statusCode,
		Message:    message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		models.LogError("Failed to encode JSON response", err)
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

func (m *ModHandler) RequestModeration(w http.ResponseWriter, r *http.Request, channelID int64) {
	ctx := r.Context()
	currentUser, ok := mw.GetUserFromContext(ctx)
	if !ok {
		models.LogWarnWithContext(ctx, "Current user not found in context for moderation request")
		return
	}
	channelOwner, err := m.App.Channels.GetNameOfChannelOwner(ctx, channelID)
	if err != nil {
		models.LogErrorWithContext(ctx, "Failed to fetch channel owner", err)
	}

	channel, err := m.App.Channels.GetChannelByID(ctx, channelID)
	if err != nil {
		models.LogErrorWithContext(ctx, "Failed to fetch channel", err)
		http.Error(w, `{"error": "channel not found"}`, http.StatusNotFound)
		return
	}

	switch channel.Privacy {
	case true:
		// construct the request, set the status to pending, notify the user
		// send a message to the channel owner
		writeJSONResponse(w, http.StatusOK, fmt.Sprintf("Moderation request sent to %s", channelOwner))
	case false:
		// call the  AddModeration function
		if m.App.Mods.AddModeration(currentUser.ID, channelID) != nil {
			models.LogErrorWithContext(ctx, "Failed to add moderation", err)
		}
		writeJSONResponse(w, http.StatusOK, fmt.Sprintf("Welcome to %s!", channel.Name))
	default:
		models.LogWarnWithContext(ctx, "Channel privacy value is neither true nor false")
	}
}
