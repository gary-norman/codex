package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gary-norman/forum/internal/app"
	"github.com/gary-norman/forum/internal/models"
)

type ReactionHandler struct {
	App *app.App
}

// GetPostsLikesAndDislikes updates the reactions of each post in the given slice
func (h *ReactionHandler) GetPostsLikesAndDislikes(posts []*models.Post) []*models.Post {
	ctx := context.Background()
	for p, post := range posts {
		likes, dislikes, err := h.App.Reactions.CountReactions(ctx, post.ID, 0) // Pass 0 for CommentID if it's a post
		// fmt.Printf("PostID: %v, Likes: %v, Dislikes: %v\n", posts[i].ID, likes, dislikes)
		if err != nil {
			models.LogError("Failed to count reactions for post", err, "PostID:", post.ID)
			likes, dislikes = 0, 0 // Default values if there is an error
		}
		models.React(posts[p], likes, dislikes)
	}
	return posts
}

func (h *ReactionHandler) getLastReactionTimeForPosts(posts []*models.Post) ([]*models.Post, error) {
	ctx := context.Background()
	for i := range posts {
		p := posts[i]

		if p.Likes < 0 || p.Dislikes < 0 {
			continue
		}

		lastReactionTime, err := h.App.Reactions.GetLastReaction(ctx, p.ID, 0)
		if err != nil {
			models.LogError("Failed to get last reaction time for post", err, "PostID:", p.ID)
		}
		if lastReactionTime.Created.IsZero() {
			// fmt.Printf("No reaction time found for PostID: %v\n", p.ID)

			p.LastReaction = nil
		} else {
			p.LastReaction = &lastReactionTime.Created
			// fmt.Printf("\nPostID: %v\nLastReaction: %v\n", p.ID, p.LastReaction)
		}
	}
	return posts, nil
}

// GetCommentsLikesAndDislikes updates the reactions of each comment in the given slice
func (h *ReactionHandler) GetCommentsLikesAndDislikes(comments []models.Comment) []models.Comment {
	ctx := context.Background()
	for c, comment := range comments {
		likes, dislikes, err := h.App.Reactions.CountReactions(ctx, 0, comment.ID) // Pass 0 for PostID if it's a comment
		// fmt.Printf("PostID: %v, Likes: %v, Dislikes: %v\n", posts[i].ID, likes, dislikes)
		if err != nil {
			models.LogError("Failed to count reactions for comment", err, "CommentID:", comment.ID)
			likes, dislikes = 0, 0 // Default values if there is an error
		}
		models.React(&comments[c], likes, dislikes)
	}
	return comments
}

func (h *ReactionHandler) StoreReaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	models.LogInfoWithContext(r.Context(), "Processing reaction storage request")

	// Variable to hold the decoded data
	var input models.ReactionInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Convert AuthorID string to UUIDField
	authorID, err := models.UUIDFieldFromString(input.AuthorID)
	if err != nil {
		http.Error(w, "Invalid authorId", http.StatusBadRequest)
		return
	}

	reactionData := models.Reaction{
		Liked:            input.Liked,
		Disliked:         input.Disliked,
		AuthorID:         authorID,
		ReactedPostID:    input.ReactedPostID,
		ReactedCommentID: input.ReactedCommentID,
	}

	//// Validate that at least one of reactedPostID or reactedCommentID is non-zero
	if (reactionData.ReactedPostID == nil || *reactionData.ReactedPostID == 0) && (reactionData.ReactedCommentID == nil || *reactionData.ReactedCommentID == 0) {
		models.LogWarnWithContext(r.Context(), "Invalid reaction data: both reactedPostID and reactedCommentID are nil or zero")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var updatedID int64
	var updatedStr string

	if reactionData.ReactedPostID != nil {
		reactionData.PostID = *reactionData.ReactedPostID
		// log.Println("ReactedPostID:", *reactionData.ReactedPostID)
		updatedID = *reactionData.ReactedPostID
		updatedStr = "post"
	} else {
		reactionData.CommentID = *reactionData.ReactedCommentID
		// log.Printf("ReactedCommentID: %d", *reactionData.ReactedPostID)
		updatedID = *reactionData.ReactedCommentID
		updatedStr = "comment"
	}

	models.LogInfoWithContext(r.Context(), "Updating reaction for %s", fmt.Sprintf("%s: %d", updatedStr, updatedID))

	if err := h.App.Reactions.Upsert(ctx, reactionData.Liked, reactionData.Disliked, reactionData.AuthorID, reactionData.PostID, reactionData.CommentID); err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to upsert reaction", err, fmt.Sprintf("%s: %d", updatedStr, updatedID))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Respond with a JSON response
	w.Header().Set("Content-Type", "application/json")
	// Send a response indicating success
	// w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "Reaction added to database"})
	if err != nil {
		models.LogErrorWithContext(r.Context(), "Failed to encode JSON response", err)
		http.Error(w, err.Error(), 500)
		return
	}

	// updatedReactions, err := h.App.Reactions.All()
	// if err != nil {
	// 	log.Printf(ErrorMsgs.Read, updatedReactions, "storeReaction > h.App.reactions.All", err)
	// 	return
	// }
	//
	// for _, reaction := range updatedReactions {
	// 	if reaction.ReactedPostID != nil {
	// 		reaction.PostID = *reaction.ReactedPostID
	// 		reaction.CommentID = 0
	// 	} else {
	// 		reaction.PostID = 0
	// 		reaction.CommentID = *reaction.ReactedCommentID
	// 	}
	// 	reaction.ReactedCommentID = nil
	// 	reaction.ReactedPostID = nil
	// 	models.JsonPost(reaction)
	// 	fmt.Println(ErrorMsgs.Divider)
	// }
}
