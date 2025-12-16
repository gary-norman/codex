package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gary-norman/forum/internal/app"
	mw "github.com/gary-norman/forum/internal/http/middleware"
	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/view"
)

type UserHandler struct {
	App      *app.App
	Reaction *ReactionHandler
	Post     *PostHandler
	Comment  *CommentHandler
	Channel  *ChannelHandler
}

func (u *UserHandler) GetThisUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Fetch User ID from the URL as string
	idStr := r.PathValue("userId")

	// Convert User ID string to FieldUUID
	userID, err := models.UUIDFieldFromString(idStr)
	if err != nil {
		view.RenderErrorPage(w, models.NotFoundLocation("user"), 400, models.NotFoundError(r.PathValue("userId"), "GetThisUser", err))
		return
	}

	userLoggedIn := true
	currentUser, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		models.LogWarnWithContext(r.Context(), "User not authenticated in GetThisUser")
		userLoggedIn = false
	}

	// if err != nil {
	// 	log.Printf(ErrorMsgs.KeyValuePair, "error parsing thisUser ID", err)
	// 	// http.Error(w, `{"error": "invalid thisUser ID"}`, http.StatusBadRequest)
	// }

	// Fetch the thisUser
	thisUser, err := u.App.Users.GetUserByID(userID)
	if err != nil {
		view.RenderErrorPage(w, models.NotFoundLocation("user"), 400, models.NotFoundError(userID, "GetThisUser", err))
	}

	// Fetch thisUser loyalty
	if err == nil {

		thisUser.Followers, thisUser.Following, err = u.App.Loyalty.CountUsers(thisUser.ID)
		if err != nil {
			view.RenderErrorPage(w, models.NotFoundLocation("user"), 500, models.FetchError("thisUser loyalty", "GetThisUser", err))
		}
	}

	// Fetch thisUser userPosts
	userPosts, err := u.App.Posts.GetPostsByUserID(thisUser.ID)
	if err != nil {
		view.RenderErrorPage(w, models.NotFoundLocation("user"), 500, models.FetchError("thisUser userPosts", "GetThisUser", err))
	}

	// Fetch Reactions for posts
	userPosts = u.Reaction.GetPostsLikesAndDislikes(userPosts)

	// Retrieve last reaction time for userPosts
	userPosts, err = u.Reaction.getLastReactionTimeForPosts(userPosts)
	if err != nil {
		view.RenderErrorPage(w, models.NotFoundLocation("user"), 500, models.FetchError("last reaction time for posts info", "GetThisUser", err))
	}

	// Fetch channel name for userPosts
	for p := range userPosts {
		userPosts[p].ChannelID, userPosts[p].ChannelName, err = u.Channel.GetChannelInfoFromPostID(userPosts[p].ID)
		if err != nil {
			view.RenderErrorPage(w, models.NotFoundLocation("user"), 500, models.FetchError("channel info", "GetThisUser", err))
		}

		models.UpdateTimeSince(userPosts[p])
	}

	// Fetch thisUser post comments
	userPosts, err = u.Comment.GetPostsComments(userPosts)
	if err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to fetch post comments", err)
	}

	models.UpdateTimeSince(&thisUser)

	// SECTION --- channels --
	allChannels, err := u.App.Channels.All()
	if err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to fetch all channels", err)
	}
	for c := range allChannels {
		models.UpdateTimeSince(allChannels[c])
	}

	for p := range userPosts {
		for _, channel := range allChannels {
			if channel.ID == userPosts[p].ChannelID {
				userPosts[p].ChannelName = channel.Name
			}
		}
	}

	ownedChannels := make([]*models.Channel, 0)
	joinedChannels := make([]*models.Channel, 0)
	ownedAndJoinedChannels := make([]*models.Channel, 0)
	channelMap := make(map[int64]bool)
	// var userPosts []models.Post

	if userLoggedIn {
		currentUser.Followers, currentUser.Following, err = u.App.Loyalty.CountUsers(currentUser.ID)
		if err != nil {
			models.LogErrorWithContext(r.Context(), "Failed to count current user loyalty", err)
		}

		// get owned and joined channels of current thisUser
		memberships, memberErr := u.App.Memberships.UserMemberships(currentUser.ID)
		if memberErr != nil {
			models.LogErrorWithContext(r.Context(), "Failed to fetch user memberships", memberErr)
		}
		ownedChannels, err = u.App.Channels.OwnedOrJoinedByCurrentUser(currentUser.ID)
		if err != nil {
			models.LogErrorWithContext(r.Context(), "Failed to fetch user owned channels", err)
		}
		joinedChannels, err = u.Channel.JoinedByCurrentUser(memberships)
		if err != nil {
			models.LogErrorWithContext(r.Context(), "Failed to fetch user joined channels", err)
		}

		// ownedAndJoinedChannels = append(ownedChannels, joinedChannels...)
		// Add owned channels
		for _, channel := range ownedChannels {
			if !channelMap[channel.ID] {
				channelMap[channel.ID] = true
				ownedAndJoinedChannels = append(ownedAndJoinedChannels, channel)
			}
		}

		// Add joined channels
		for _, channel := range joinedChannels {
			if !channelMap[channel.ID] {
				channelMap[channel.ID] = true
				ownedAndJoinedChannels = append(ownedAndJoinedChannels, channel)
			}
		}
	} else {
		ownedAndJoinedChannels = allChannels
	}

	data := models.UserPage{
		UserID:      models.NewUUIDField(), // Default value of 0 for logged out users
		CurrentUser: currentUser,
		Instance:    "user-page",
		ThisUser:    &thisUser,
		ImagePaths:  u.App.Paths,
		// ---------- userPosts ----------
		Posts: userPosts,
		// ---------- channels ----------
		AllChannels:            allChannels,
		OwnedChannels:          ownedChannels,
		JoinedChannels:         joinedChannels,
		OwnedAndJoinedChannels: ownedAndJoinedChannels,
	}

	view.RenderPageData(w, data)
}

// GetLoggedInUser gets the currently logged-in user from the session token and returns the user's struct
func (u *UserHandler) GetLoggedInUser(r *http.Request) (*models.User, error) {
	// Get the username from the request cookie
	userCookie, getCookieErr := r.Cookie("username")
	if getCookieErr != nil {
		return nil, fmt.Errorf("failed to get username cookie: %w", getCookieErr)
	}
	var username string
	if userCookie != nil {
		username = userCookie.Value
	}
	models.LogInfo("Retrieved username from cookie: %s", username)
	if username == "" {
		return nil, errors.New("no user is logged in")
	}
	user, getUserErr := u.App.Users.GetUserByUsername(username, "GetLoggedInUser")
	if getUserErr != nil {
		return nil, getUserErr
	}
	return user, nil
}

func (u *UserHandler) EditUserDetails(w http.ResponseWriter, r *http.Request) {
	user, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		models.LogErrorWithContext(r.Context(), "User not found in context for EditUserDetails", nil)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to parse multipart form in EditUserDetails", err)
		http.Error(w, err.Error(), 400)
		return
	}
	currentAvatar := user.Avatar
	prefix := "noimage"
	models.LogInfoWithContext(r.Context(), "Current avatar: %v", currentAvatar)
	user.Avatar = GetFileName(r, "file-drop", "editUserDetails", "user")
	// TODO does this check need to be here?
	if strings.HasPrefix(currentAvatar, prefix) {
		user.Avatar = currentAvatar
	}
	currentDescription := r.FormValue("bio")
	if currentDescription != "" {
		user.Description = currentDescription
	}
	currentName := r.FormValue("name")
	if currentName != "" {
		user.Username = currentName
	}
	editErr := u.App.Users.Edit(user)
	if editErr != nil {
		models.LogErrorWithContext(r.Context(), "Failed to edit user details", editErr)
	}
	ephemeral := true
	if err, _ := u.App.Cookies.CreateCookies(w, user, ephemeral); err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to create cookies", err)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
