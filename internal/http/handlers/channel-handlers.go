package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gary-norman/forum/internal/app"
	mw "github.com/gary-norman/forum/internal/http/middleware"
	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/view"
)

type ChannelHandler struct {
	App      *app.App
	Reaction *ReactionHandler
	Comment  *CommentHandler
}

func (c *ChannelHandler) GetChannelPage(w http.ResponseWriter, r *http.Request) {
	models.LogInfoWithContext(r.Context(), "GetChannelPage called")
	w.Header().Set("Content-Type", "application/json")
	userLoggedIn := true
	currentUser, ok := mw.GetUserFromContext(r.Context())

	if !ok {
		models.LogWarnWithContext(r.Context(), "User not authenticated in GetChannelPage")
		userLoggedIn = false
	}

	allChannels, err := c.App.Channels.All()
	if err != nil {
		http.Error(w, `{"error": "Error getting all channels"}`, http.StatusInternalServerError)
	}
	for c := range allChannels {
		models.UpdateTimeSince(allChannels[c])
	}
	// get owner name and timesince
	// get owned channels
	// get joined channels
	// innerjoin these 2

	ownedChannels := make([]*models.Channel, 0)
	joinedChannels := make([]*models.Channel, 0)
	ownedAndJoinedChannels := make([]*models.Channel, 0)

	if userLoggedIn {
		// get owned and joined channels of current user
		memberships, memberErr := c.App.Memberships.UserMemberships(currentUser.ID)
		if memberErr != nil {
			models.LogErrorWithContext(r.Context(), "Failed to fetch user memberships", memberErr)
			http.Error(w, `{"error": "Error getting user memberships"}`, http.StatusInternalServerError)
		}
		ownedChannels, err = c.App.Channels.OwnedOrJoinedByCurrentUser(currentUser.ID)
		if err != nil {
			models.LogErrorWithContext(r.Context(), "Failed to fetch user owned channels", err)
			http.Error(w, `{"error": "Error getting user owned channels"}`, http.StatusInternalServerError)
		}
		joinedChannels, err = c.JoinedByCurrentUser(memberships)
		if err != nil {
			models.LogErrorWithContext(r.Context(), "Failed to fetch user joined channels", err)
			http.Error(w, `{"error": "Error getting user joined channels"}`, http.StatusInternalServerError)
		}
		ownedAndJoinedChannels = append(ownedChannels, joinedChannels...)
	}

	// TODO make a better struct for all
	data := models.ChannelPage{
		UserID:                 models.NewUUIDField(), // Default value of 0 for logged out users
		CurrentUser:            currentUser,
		Instance:               "all-channels-page",
		OwnedChannels:          ownedChannels,
		JoinedChannels:         joinedChannels,
		OwnedAndJoinedChannels: ownedAndJoinedChannels,
		ImagePaths:             c.App.Paths,
	}
	view.RenderPageData(w, data)
}

func (c *ChannelHandler) GetThisChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	models.LogInfoWithContext(r.Context(), "GetThisChannel called")
	userLoggedIn := true
	currentUser, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		models.LogWarnWithContext(r.Context(), "User not authenticated in GetThisChannel")
		userLoggedIn = false
	}

	// Parse channelID from the request
	channelID, err := models.GetIntFromPathValue(r.PathValue("channelId"))
	if err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to parse channel ID from path value", err)
		error := fmt.Errorf("channel not found: %s: %w", r.PathValue("channelId"), err)
		view.RenderErrorPage(w, models.NotFoundLocation("channel"), 400, error)
		return
	}

	// Fetch the channel
	foundChannels, err := c.App.Channels.GetChannelsByID(channelID)
	if err != nil || len(foundChannels) == 0 {
		models.LogErrorWithContext(r.Context(), "Channel not found by ID", err)
		error := fmt.Errorf("channel %d not found: %w", channelID, err)
		view.RenderErrorPage(w, models.NotFoundLocation("channel"), 400, error)
		return
	}
	thisChannel := foundChannels[0]
	models.LogInfoWithContext(r.Context(), "Fetching channel: %s", thisChannel.Name)
	models.UpdateTimeSince(thisChannel)

	// Fetch the channel owner
	thisChannelOwnerName, ownerErr := c.App.Users.GetSingleUserValue(thisChannel.OwnerID, "ID", "username")
	if ownerErr != nil {
		http.Error(w, `{"error": "Error getting channel owner"}`, http.StatusInternalServerError)
		return
	}

	// Fetch channel rules
	thisChannelRules, err := c.App.Rules.AllForChannel(thisChannel.ID)
	if err != nil {
		http.Error(w, `{"error": "Error getting channel rules"}`, http.StatusInternalServerError)
		return
	}

	// Fetch channel posts
	var thisChannelPosts []*models.Post
	thisChannelPostIDs, err := c.App.Channels.GetPostIDsFromChannel(thisChannel.ID)
	if err != nil {
		http.Error(w, `{"error": "Error getting Post IDs"}`, http.StatusInternalServerError)
	}
	for p := range thisChannelPostIDs {
		post, err := c.App.Posts.GetPostByID(thisChannelPostIDs[p])
		if err != nil {
			http.Error(w, `{"error": "Error getting post ID:" + thisChannelPostIDs[p]}`, http.StatusInternalServerError)
		}
		thisChannelPosts = append(thisChannelPosts, &post)
	}

	allChannels, err := c.App.Channels.All()
	if err != nil {
		http.Error(w, `{"error": "Error getting all channels"}`, http.StatusInternalServerError)
	}
	for c := range allChannels {
		models.UpdateTimeSince(allChannels[c])
	}
	channelName, err := c.App.Channels.GetChannelNameFromID(thisChannel.ID)
	if err != nil {
		http.Error(w, `{"error": "error fetching channel name"}`, http.StatusInternalServerError)
	}

	// Add channel ID & name and fetch timesince for posts
	for p := range thisChannelPosts {
		thisChannelPosts[p].ChannelID, thisChannelPosts[p].ChannelName = thisChannel.ID, channelName
		// TODO no need for this
		models.UpdateTimeSince(thisChannelPosts[p])
	}

	// Retrieve total likes and dislikes for each Channel post
	thisChannelPosts = c.Reaction.GetPostsLikesAndDislikes(thisChannelPosts)

	// Retrieve last reaction time for posts
	thisChannelPosts, err = c.Reaction.getLastReactionTimeForPosts(thisChannelPosts)
	if err != nil {
		http.Error(w, `{"error": "Error getting channel posts" }`, http.StatusInternalServerError)
	}

	// Retrieve comments for posts
	thisChannelPosts, err = c.Comment.GetPostsComments(thisChannelPosts)
	if err != nil {
		http.Error(w, `{"error": "Error getting comments" }`, http.StatusInternalServerError)
	}

	ownedChannels := make([]*models.Channel, 0)
	joinedChannels := make([]*models.Channel, 0)
	ownedAndJoinedChannels := make([]*models.Channel, 0)
	isJoined := false
	isOwned := false

	if userLoggedIn {
		isOwned = currentUser.ID == thisChannel.OwnerID
		// attach following/follower numbers to currently logged-in user
		currentUser.Followers, currentUser.Following, err = c.App.Loyalty.CountUsers(currentUser.ID)
		if err != nil {
			models.LogErrorWithContext(r.Context(), "Failed to count current user loyalty", err)
			http.Error(w, `{"error": "Error getting user loyalty"}`, http.StatusInternalServerError)
		}
		// get owned and joined channels of current user
		memberships, memberErr := c.App.Memberships.UserMemberships(currentUser.ID)
		if memberErr != nil {
			models.LogErrorWithContext(r.Context(), "Failed to fetch user memberships", memberErr)
			http.Error(w, `{"error": "Error getting user memberships"}`, http.StatusInternalServerError)
		}
		ownedChannels, err = c.App.Channels.OwnedOrJoinedByCurrentUser(currentUser.ID)
		if err != nil {
			models.LogErrorWithContext(r.Context(), "Failed to fetch user owned channels", err)
			http.Error(w, `{"error": "Error getting user owned channels"}`, http.StatusInternalServerError)
		}
		joinedChannels, err = c.JoinedByCurrentUser(memberships)
		if err != nil {
			models.LogErrorWithContext(r.Context(), "Failed to fetch user joined channels", err)
			http.Error(w, `{"error": "Error getting user joined channels"}`, http.StatusInternalServerError)
		}
		ownedAndJoinedChannels = append(ownedChannels, joinedChannels...)

		// Determine whether the current user has joined thisChannel
		for _, channel := range joinedChannels {
			if thisChannel.ID == channel.ID {
				isJoined = true
				break
			}
		}
	}

	data := models.ChannelPage{
		UserID:                 models.NewUUIDField(), // Default value of 0 for logged out users
		CurrentUser:            currentUser,
		Instance:               "channel-page",
		ThisChannel:            thisChannel,
		OwnerName:              thisChannelOwnerName,
		IsOwned:                isOwned,
		IsJoined:               isJoined,
		Rules:                  thisChannelRules,
		Posts:                  thisChannelPosts,
		OwnedChannels:          ownedChannels,
		JoinedChannels:         joinedChannels,
		OwnedAndJoinedChannels: ownedAndJoinedChannels,
		ImagePaths:             c.App.Paths,
	}
	view.RenderPageData(w, data)
}

func (c *ChannelHandler) GetChannelInfoFromPostID(postID int64) (int64, string, error) {
	channelIDs, err := c.App.Channels.GetChannelIDFromPost(postID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get channel ID from post %d: %w", postID, err)
	}
	channelID := channelIDs[0]
	channelName, err := c.App.Channels.GetChannelNameFromID(channelID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get channel name from ID %d: %w", channelID, err)
	}
	return channelID, channelName, nil
}

func (c *ChannelHandler) StoreChannel(w http.ResponseWriter, r *http.Request) {
	user, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	parseErr := r.ParseMultipartForm(10 << 20)
	if parseErr != nil {
		models.LogErrorWithContext(r.Context(), "Failed to parse multipart form in StoreChannel", parseErr)
		http.Error(w, parseErr.Error(), 400)
		return
	}

	createChannelData := models.Channel{
		OwnerID:     user.ID,
		Name:        r.PostForm.Get("name"),
		Description: r.PostForm.Get("description"),
		Avatar:      "noimage",
		Banner:      "default.png",
		Privacy:     false,
		IsFlagged:   false,
		IsMuted:     false,
	}
	if r.PostForm.Get("privacy") == "on" {
		createChannelData.Privacy = true
	}
	createChannelData.Avatar = GetFileName(r, "file-drop", "storeChannel", "channel")

	insertErr := c.App.Channels.Insert(
		createChannelData.OwnerID,
		createChannelData.Name,
		createChannelData.Description,
		createChannelData.Avatar,
		createChannelData.Banner,
		createChannelData.Privacy,
		createChannelData.IsFlagged,
		createChannelData.IsMuted,
	)

	if insertErr != nil {
		models.LogErrorWithContext(r.Context(), "Failed to insert channel", insertErr)
		http.Error(w, insertErr.Error(), 500)
		return
	}
	// TODO fix this redirect
	http.Redirect(w, r, "/", http.StatusFound)
}

func (c *ChannelHandler) StoreMembership(w http.ResponseWriter, r *http.Request) {
	user, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if parseErr := r.ParseForm(); parseErr != nil {
		models.LogErrorWithContext(r.Context(), "Failed to parse form in StoreMembership", parseErr)
		http.Error(w, parseErr.Error(), 400)
		return
	}
	models.LogInfoWithContext(r.Context(), "User %s joining channel", user.Username)
	// get channelID
	joinedChannelID, convErr := strconv.ParseInt(r.PostForm.Get("channelId"), 10, 64)
	if convErr != nil {
		models.LogErrorWithContext(r.Context(), "Failed to convert channelId to int", convErr)
	}
	if err := c.App.Memberships.Insert(user.ID, joinedChannelID); err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to insert membership", err)
		http.Error(w, err.Error(), 500)
		return
	}

	channelName, err := c.App.Channels.GetNameOfChannel(joinedChannelID)
	if err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to get channel name", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encErr := json.NewEncoder(w).Encode(map[string]any{
		"code":    http.StatusOK,
		"message": fmt.Sprintf("Welcome to %v!", channelName),
	})
	if encErr != nil {
		models.LogErrorWithContext(r.Context(), "Failed to encode response in StoreMembership", encErr)
		return
	}
}

// TODO this function is a mess

// JoinedByCurrentUser checks if the currently logged-in user is a member of the current channel
func (c *ChannelHandler) JoinedByCurrentUser(memberships []models.Membership) ([]*models.Channel, error) {
	models.LogInfo("Checking user memberships")
	var channels []*models.Channel
	for _, membership := range memberships {
		channel, err := c.App.Channels.GetChannelsByID(membership.ChannelID)
		if err != nil {
			return nil, fmt.Errorf("failed to get channel %d for membership: %w", membership.ChannelID, err)
		}
		channels = append(channels, channel[0])
	}
	// TODO add logic that checks if the user is an owner of this channel
	if len(channels) > 0 {
		models.LogInfo("User is a member of %d channels", len(channels))
	} else {
		models.LogInfo("User is not a member of any channels")
	}
	return channels, nil
}

func (c *ChannelHandler) CreateAndInsertRule(w http.ResponseWriter, r *http.Request) {
	channelID, err := strconv.ParseInt(r.PathValue("channelId"), 10, 64)
	if err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to convert channelId to int", err)
	}

	// Get the "rules" input value
	rulesJSON := r.FormValue("rules")
	if rulesJSON == "" { // TODO send this message to the user
		models.LogWarnWithContext(r.Context(), "No rules added or removed by user")
	}

	// Decode JSON into a slice of Rule structs
	var rules []models.PostRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	for _, rule := range rules {
		id, found := strings.CutPrefix(rule.ID, "existing-channel-rule-")
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			models.LogErrorWithContext(r.Context(), "Failed to convert rule ID to int", err)
		}
		if found {
			err := c.App.Rules.DeleteRule(channelID, idInt)
			if err != nil {
				models.LogErrorWithContext(r.Context(), "Failed to delete rule", err)
			}
		} else {
			ruleID, err := c.App.Rules.CreateRule(rule.Rule)
			if err != nil {
				models.LogErrorWithContext(r.Context(), "Failed to create rule", err)
			}
			err = c.App.Rules.InsertRule(channelID, ruleID)
			if err != nil {
				models.LogErrorWithContext(r.Context(), "Failed to insert rule", err)
			}
		}
	}
	http.Redirect(w, r, "/channels/"+r.PathValue("channelId"), http.StatusFound)
}
