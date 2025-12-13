package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gary-norman/forum/internal/app"
	mw "github.com/gary-norman/forum/internal/http/middleware"
	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/view"
)

type PostHandler struct {
	App      *app.App
	Channel  *ChannelHandler
	Comment  *CommentHandler
	Reaction *ReactionHandler
}

// GetUserPosts returns a slice of Posts that belong to channels the user follows. If no user is logged in, it returns all posts
func (p *PostHandler) GetUserPosts(user *models.User, allPosts []*models.Post) []*models.Post {
	if user == nil {
		return allPosts
	}

	memberships, memberErr := p.App.Memberships.UserMemberships(user.ID)
	if memberErr != nil {
		log.Printf(ErrorMsgs.KeyValuePair, "GetUserPosts > UserMemberships", memberErr)
		return allPosts
	}
	ownedChannels, ownedErr := p.App.Channels.OwnedOrJoinedByCurrentUser(user.ID)
	if ownedErr != nil {
		log.Printf(ErrorMsgs.Query, "GetUserPosts > user owned channels", ownedErr)
		return allPosts
	}
	joinedChannels, joinedErr := p.Channel.JoinedByCurrentUser(memberships)
	if joinedErr != nil {
		log.Printf(ErrorMsgs.Query, "user joined channels", joinedErr)
		return allPosts
	}

	// Combine owned and joined channels
	allChannels := make([]*models.Channel, 0, len(ownedChannels)+len(joinedChannels))
	allChannels = append(allChannels, ownedChannels...)
	allChannels = append(allChannels, joinedChannels...)

	if len(allChannels) == 0 {
		return allPosts
	}

	// Build a set of channel IDs for fast lookup
	channelIDSet := make(map[int64]struct{}, len(allChannels))
	for _, ch := range allChannels {
		channelIDSet[ch.ID] = struct{}{}
	}

	// Filter posts belonging to user's channels
	postsInUserChannels := make([]*models.Post, 0, len(allPosts))
	for _, post := range allPosts {
		if _, exists := channelIDSet[post.ChannelID]; exists {
			postsInUserChannels = append(postsInUserChannels, post)
		}
	}

	return postsInUserChannels
}

func (p *PostHandler) GetThisPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var thisPost *models.Post
	var posts []*models.Post
	isMember := false
	var isMemberErr error
	isOwner := false

	userLoggedIn := true
	currentUser, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		fmt.Println(ErrorMsgs.NotFound, "currentUser", "getThisPost", "_")
		userLoggedIn = false
	}

	// Parse post ID from URL
	postID, err := models.GetIntFromPathValue(r.PathValue("postId"))
	if err != nil {
		view.RenderErrorPage(w, models.NotFoundLocation("post"), 400, models.NotFoundError(r.PathValue("postId"), "GetThisPost", err))
		return
	}

	// Fetch the post
	post, err := p.App.Posts.GetPostByID(postID)
	if err != nil {
		view.RenderErrorPage(w, models.NotFoundLocation("post"), 400, models.NotFoundError(postID, "GetThisPost", err))
		return
	}
	posts = append(posts, &post)
	foundPosts, err := p.Comment.GetPostsComments(posts)
	if err != nil {
		view.RenderErrorPage(w, models.NotFoundLocation("post"), 500, models.FetchError("post comments", "GetThisPost", err))
	}
	foundPosts = p.Reaction.GetPostsLikesAndDislikes(foundPosts)
	thisPost = foundPosts[0]

	thisPost.ChannelID, thisPost.ChannelName, err = p.Channel.GetChannelInfoFromPostID(thisPost.ID)
	if err != nil {
		view.RenderErrorPage(w, models.NotFoundLocation("post"), 500, models.FetchError("channel info", "GetThisPost", err))
	}

	models.UpdateTimeSince(thisPost)

	// Fetch the channel
	channel, err := p.App.Channels.GetChannelByID(thisPost.ChannelID)
	if err != nil {
		view.RenderErrorPage(w, models.NotFoundLocation("post"), 500, models.QueryError("channels", "GetThisPost", err))
	}

	// Fetch the author
	author, err := p.App.Users.GetUserByUsername(thisPost.Author, "GetThisPost")
	if err != nil {
		view.RenderErrorPage(w, models.NotFoundLocation("post"), 500, models.QueryError("users", "GetThisPost", err))
	}

	if userLoggedIn {
		currentUser.Followers, currentUser.Following, err = p.App.Loyalty.CountUsers(currentUser.ID)
		if err != nil {
			view.RenderErrorPage(w, models.NotFoundLocation("post"), 500, models.FetchError("user loyalty", "GetThisPost", err))
		}
		// Fetch if the user is a member of the channel
		isMember, isMemberErr = p.App.Channels.IsUserMemberOfChannel(currentUser.ID, channel.ID)
		if isMemberErr != nil {
			view.RenderErrorPage(w, models.NotFoundLocation("post"), 500, models.QueryError("user membership", "GetThisPost", err))
		}
		// Fetch if the user is the owner of the channel
		isOwner = currentUser.ID == channel.OwnerID
	}

	data := models.PostPage{
		UserID:      models.NewUUIDField(), // Default value of 0 for logged out users
		CurrentUser: currentUser,
		Instance:    "post-page",
		ThisPost:    thisPost,
		Author:      author,
		ThisChannel: channel,
		OwnerName:   author.Username,
		IsOwned:     isOwner,
		IsJoined:    isMember,
		ImagePaths:  p.App.Paths,
	}
	view.RenderPageData(w, data)
}

func (p *PostHandler) StorePost(w http.ResponseWriter, r *http.Request) {
	user, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf(ErrorMsgs.Parse, "storePost", err)
		return
	}

	channels := r.MultipartForm.Value["post_channel_list"]
	if len(channels) < 1 {
		log.Println(fmt.Errorf("at least one channel required"))
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	if title == "" || content == "" {
		log.Println(fmt.Errorf("title and content are required"))
		return
	}

	createPostData := models.Post{
		Title:         title,
		Content:       content,
		Author:        user.Username,
		AuthorID:      user.ID,
		AuthorAvatar:  user.Avatar,
		IsCommentable: r.FormValue("commentable") == "on",
	}

	if img := GetFileName(r, "file-drop", "storePost", "post"); img != "" {
		createPostData.Images = img
	}

	postID, err := p.App.Posts.Insert(
		createPostData.Title,
		createPostData.Content,
		createPostData.Images,
		createPostData.Author,
		createPostData.AuthorAvatar,
		createPostData.AuthorID,
		createPostData.IsCommentable,
		createPostData.IsFlagged,
	)
	if err != nil {
		log.Printf(ErrorMsgs.Post, err)
		//TODO: internal server errors should be handled better and redirect to NotFound page
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, c := range channels {
		channelID, convErr := strconv.ParseInt(c, 10, 64)
		if convErr != nil {
			http.Error(w, "invalid channel id", http.StatusBadRequest)
			return
		}
		if err := p.App.Channels.AddPostToChannel(channelID, postID); err != nil {
			log.Printf("failed adding post %d to channel %d: %v", postID, channelID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// ✅ Redirect only — no JSON write
	postURL := fmt.Sprintf("/cdx/post/%d", postID)
	http.Redirect(w, r, postURL, http.StatusSeeOther)
}

// SECTION getting channel data (for reverting to single channel post)

//selectionJSON := r.PostForm.Get("channel")
//if selectionJSON == "" {
//	http.Error(w, "No selection provided", http.StatusBadRequest)
//	return
//}
//var channelData models.ChannelData
//if err := json.Unmarshal([]byte(selectionJSON), &channelData); err != nil {
//	log.Printf(ErrorMsgs.Unmarshal, selectionJSON, err)
//	http.Error(w, "Invalid selection format", http.StatusBadRequest)
//	return
//}
//fmt.Printf(ErrorMsgs.KeyValuePair, "channelName", channelData.ChannelName)
//fmt.Printf(ErrorMsgs.KeyValuePair, "channelID", channelData.ChannelID)
//fmt.Printf(ErrorMsgs.KeyValuePair, "commentable", r.PostForm.Get("commentable"))
