package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gary-norman/forum/internal/app"
	mw "github.com/gary-norman/forum/internal/http/middleware"
	"github.com/gary-norman/forum/internal/models"
)

type ChatHandler struct {
	App *app.App
}

// CreateChat creates a new buddy chat between current user and another user
func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get current user
	currentUser, ok := mw.GetUserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		BuddyID string `json:"buddy_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.LogErrorWithContext(ctx, "Failed to decode create chat request", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Parse buddy ID
	var buddyUUID models.UUIDField
	if err := buddyUUID.UnmarshalJSON([]byte(`"` + req.BuddyID + `"`)); err != nil {
		models.LogErrorWithContext(ctx, "Invalid buddy ID format", err)
		http.Error(w, "Invalid buddy ID", http.StatusBadRequest)
		return
	}

	// Check if chat already exists between these users
	existingChatID, err := h.App.Chats.GetBuddyChatID(ctx, currentUser.ID, buddyUUID)
	if err == nil && existingChatID.UUID.String() != "00000000-0000-0000-0000-000000000000" {
		// Chat already exists, return it
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"chat_id": existingChatID.String(),
			"exists":  true,
		})
		return
	}

	// Create new chat (for buddy chat: groupID should be NULL, buddyID should have value)
	// CreateChat params: (ctx, chatType, name, groupID, buddyID)
	// For buddy chats: groupID is NullableUUIDField with Valid=false (SQL NULL), buddyID has the buddy's UUID
	chatID, err := h.App.Chats.CreateChat(ctx, "buddy", "",
		models.NullableUUIDField{Valid: false}, // groupID = NULL
		models.NullableUUIDField{UUID: buddyUUID, Valid: true}) // buddyID = actual UUID
	if err != nil {
		models.LogErrorWithContext(ctx, "Failed to create chat", err)
		http.Error(w, "Failed to create chat", http.StatusInternalServerError)
		return
	}

	// Attach both users to the chat
	if err := h.App.Chats.AttachUserToChat(ctx, chatID, currentUser.ID); err != nil {
		models.LogErrorWithContext(ctx, "Failed to attach current user to chat", err)
		http.Error(w, "Failed to create chat", http.StatusInternalServerError)
		return
	}

	if err := h.App.Chats.AttachUserToChat(ctx, chatID, buddyUUID); err != nil {
		models.LogErrorWithContext(ctx, "Failed to attach buddy to chat", err)
		http.Error(w, "Failed to create chat", http.StatusInternalServerError)
		return
	}

	models.LogInfoWithContext(ctx, "Chat created between %s and %s", currentUser.Username, req.BuddyID)

	// Return chat ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"chat_id": chatID.String(),
		"exists":  false,
	})
}
