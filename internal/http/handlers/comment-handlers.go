package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gary-norman/forum/internal/app"
	mw "github.com/gary-norman/forum/internal/http/middleware"
	"github.com/gary-norman/forum/internal/models"
)

type CommentHandler struct {
	App      *app.App
	Reaction *ReactionHandler
}

func (h *CommentHandler) StoreComment(w http.ResponseWriter, r *http.Request) {
	// SECTION getting user
	user, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "No user found in context", http.StatusUnauthorized)
		return
	}

	// SECTION retrieving comment form data
	// Variable to hold the decoded data
	var commentData models.Comment

	parseErr := r.ParseMultipartForm(10 << 20)
	if parseErr != nil {
		http.Error(w, parseErr.Error(), 400)
		models.LogErrorWithContext(r.Context(), "Failed to parse multipart form", parseErr)
		return
	}

	// SECTION setting channel data
	// Get channel data
	selectionJSON := r.PostForm.Get("channel")
	if selectionJSON == "" {
		http.Error(w, "No selection provided for channel", http.StatusBadRequest)
		return
	}

	var channelData models.ChannelData
	if err := json.Unmarshal([]byte(selectionJSON), &channelData); err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to unmarshal channel data", err)
		http.Error(w, "Invalid selection format", http.StatusBadRequest)
		return
	}

	postIDStr := r.PostForm.Get("postID")
	commentIDStr := r.PostForm.Get("commentID")

	// Convert postIDStr to an integer
	postID, postConvErr := strconv.ParseInt(postIDStr, 10, 64)
	if postConvErr != nil {
		models.LogWarnWithContext(r.Context(), "Failed to convert postID: %s", postConvErr, postIDStr)
	}

	// Convert commentIDStr to an integer
	commentID, commentConvErr := strconv.ParseInt(commentIDStr, 10, 64)
	if commentConvErr != nil {
		models.LogWarnWithContext(r.Context(), "Failed to convert commentID: %s", commentConvErr, commentIDStr)
	}

	// Convert ChannelID to an integer
	channelID, channelIDConvErr := strconv.ParseInt(channelData.ChannelID, 10, 64)
	if channelIDConvErr != nil {
		models.LogWarnWithContext(r.Context(), "Failed to convert channelID: %s", channelIDConvErr, channelData.ChannelID)
	}

	// Assign the returned values
	commentData = models.Comment{
		Content:       r.PostForm.Get("content"),
		Author:        user.Username,
		AuthorID:      user.ID,
		AuthorAvatar:  user.Avatar,
		ChannelID:     channelID,
		ChannelName:   channelData.ChannelName,
		IsCommentable: true,
		IsReply:       false,
		IsFlagged:     false,
	}
	// Set CommentedPostID if it is provided (ie not 0)
	if postID != 0 {
		commentData.CommentedPostID = sql.NullInt64{Int64: int64(postID), Valid: true}
	}

	// Set CommentedCommentID if it is provided (ie not 0)
	if commentID != 0 {
		commentData.CommentedCommentID = sql.NullInt64{Int64: int64(commentID), Valid: true}
	}

	commentData.ChannelID, _ = strconv.ParseInt(channelData.ChannelID, 10, 64)

	// Log the values
	fmt.Printf("commentData.CommentedPostID: %v\n", commentData.CommentedPostID)
	fmt.Printf("commentData.CommentedCommentID: %v\n", commentData.CommentedCommentID)

	// Insert the comment
	insertErr := h.App.Comments.Upsert(commentData)

	if insertErr != nil {
		models.LogErrorWithContext(r.Context(), "Failed to upsert comment", insertErr)
		http.Error(w, insertErr.Error(), 500)
		return
	}

	path := strings.TrimSuffix(r.URL.Path, "/store-comment")

	http.Redirect(w, r, path, http.StatusFound)
}

func (h *CommentHandler) GetPostsComments(posts []*models.Post) ([]*models.Post, error) {
	for p, post := range posts {
		comments, err := h.App.Comments.GetCommentByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		comments = h.Reaction.GetCommentsLikesAndDislikes(comments)
		/// Filter comments that belong to the current post based on the postID and CommentedPostID
		var postComments []models.Comment
		var commentsCount int
		for _, comment := range comments {
			models.UpdateTimeSince(&comment)
			// For each comment, recursively assign its replies
			commentWithReplies := h.GetRepliesForComment(comment)
			postComments = append(postComments, commentWithReplies)
		}
		commentsCount = len(postComments)
		posts[p].Comments = postComments
		posts[p].CommentsCount = commentsCount
	}
	return posts, nil
}

// GetRepliesForComment Recursively fetches replies for each comment
func (h *CommentHandler) GetRepliesForComment(comment models.Comment) models.Comment {
	// Find replies to the current comment
	var replies []models.Comment
	comments, _ := h.App.Comments.GetCommentByCommentID(comment.ID)
	comments = h.Reaction.GetCommentsLikesAndDislikes(comments)
	for r, possibleReply := range comments {
		models.UpdateTimeSince(&comments[r])
		replyWithReplies := h.GetRepliesForComment(possibleReply) // Recursively get replies for this reply
		replies = append(replies, replyWithReplies)
	}
	if len(replies) > 0 {
		comment.Replies = replies
	}
	// Return the comment with its replies
	return comment
}
