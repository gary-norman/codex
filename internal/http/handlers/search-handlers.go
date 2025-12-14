package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gary-norman/forum/internal/app"
	mw "github.com/gary-norman/forum/internal/http/middleware"
)

type SearchHandler struct {
	App *app.App
}

func (s *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// SECTION --- channels --
	allChannels, err := s.App.Channels.All()
	if err != nil {
		log.Printf(ErrorMsgs.Query, "channels.All-search", err)
	}

	// SECTION --- posts ---
	// var userLoggedIn bool
	allPosts, err := s.App.Posts.All()
	if err != nil {
		log.Printf(ErrorMsgs.KeyValuePair, "Error fetching all posts", err)
	}

	for p := range allPosts {
		channelIDs, err := s.App.Channels.GetChannelIDFromPost(allPosts[p].ID)
		if err != nil {
			log.Printf(ErrorMsgs.KeyValuePair, "getHome > channelID", err)
			return
		}
		allPosts[p].ChannelID = channelIDs[0]
	}

	for p := range allPosts {
		for _, channel := range allChannels {
			if channel.ID == allPosts[p].ChannelID {
				allPosts[p].ChannelName = channel.Name
			}
		}
	}

	// SECTION --- user ---
	allUsers, allUsersErr := s.App.Users.All()
	if allUsersErr != nil {
		log.Printf(ErrorMsgs.Query, "getHome> users > All", allUsersErr)
	}

	//log.Printf("allUsers: %v\n\n", allUsers)

	currentUser, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		log.Printf(ErrorMsgs.KeyValuePair, "User is not logged in. CurrentUser", currentUser)
	}

	searchResults := map[string]any{
		"users":    allUsers,
		"channels": allChannels,
		"posts":    allPosts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(searchResults); err != nil {
		log.Printf(ErrorMsgs.Encode, "search results", err)
		http.Error(w, "Error encoding search results", http.StatusInternalServerError)
		return
	}

	log.Printf(ErrorMsgs.KeyValuePair, "Search data fetched in", time.Since(start))
}
