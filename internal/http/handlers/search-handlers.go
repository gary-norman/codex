package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gary-norman/forum/internal/app"
	mw "github.com/gary-norman/forum/internal/http/middleware"
	"github.com/gary-norman/forum/internal/models"
)

type SearchHandler struct {
	App *app.App
}

func (s *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	// Use concurrent search with request context
	result, err := ConcurrentSearch(r.Context(), s.App)
	if err != nil {
		models.LogWarnWithContext(r.Context(), "Search completed with errors: %v", err)
	}

	// Enrich posts with channel information
	enrichedPosts := enrichPostsWithChannels(s.App, result.Posts, result.Channels)

	currentUser, ok := mw.GetUserFromContext(r.Context())

	if !ok {
		models.LogInfoWithContext(r.Context(), "Anonymous user accessing search")
	} else {
		models.LogInfoWithContext(r.Context(), "User %s accessing search", currentUser.ID)
	}

	searchResults := map[string]any{
		"users":    result.Users,
		"channels": result.Channels,
		"posts":    enrichedPosts,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(searchResults); err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to encode search results", err)
		http.Error(w, "Error encoding search results", http.StatusInternalServerError)
		return
	}
}
