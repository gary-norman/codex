package handlers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gary-norman/forum/internal/app"
	"github.com/gary-norman/forum/internal/models"
)

// SearchResult holds aggregated search results from multiple sources
type SearchResult struct {
	Users    []*models.User
	Posts    []*models.Post
	Channels []*models.Channel
	Errors   []error // Collect errors from goroutines
	Duration time.Duration
}

// searchError wraps errors with source information
type searchError struct {
	Source string
	Err    error
}

func (e searchError) Error() string {
	return fmt.Sprintf("%s: %v", e.Source, e.Err)
}

// ConcurrentSearch performs parallel search across users, posts, and channels
// Uses fan-out pattern to execute queries concurrently, then fan-in results
func ConcurrentSearch(ctx context.Context, app *app.App) (*SearchResult, error) {
	start := time.Now()

	// Create result channels for each search type
	usersCh := make(chan []*models.User, 1)
	postsCh := make(chan []*models.Post, 1)
	channelsCh := make(chan []*models.Channel, 1)
	errorsCh := make(chan searchError, 3) // Buffer for 3 possible errors

	// WaitGroup to track goroutine completion
	var wg sync.WaitGroup

	// TODO(human): Exercise 4 Part 1 - Implement concurrent user search
	// Launch goroutine to search users
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Check for context cancellation
		select {
		case <-ctx.Done():
			errorsCh <- searchError{Source: "users", Err: ctx.Err()}
			return
		default:
		}

		// Perform user search
		users, err := app.Users.All()
		if err != nil {
			errorsCh <- searchError{Source: "users", Err: err}
			return
		}
		usersCh <- users
	}()

	// TODO(human): Exercise 4 Part 2 - Implement concurrent post search
	// Launch goroutine to search posts
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Check for context cancellation
		select {
		case <-ctx.Done():
			errorsCh <- searchError{Source: "posts", Err: ctx.Err()}
			return
		default:
		}

		// Perform post search
		posts, err := app.Posts.All()
		if err != nil {
			errorsCh <- searchError{Source: "posts", Err: err}
			return
		}
		postsCh <- posts
	}()

	// TODO(human): Exercise 4 Part 3 - Implement concurrent channel search
	// Launch goroutine to search channels
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Check for context cancellation
		select {
		case <-ctx.Done():
			errorsCh <- searchError{Source: "channels", Err: ctx.Err()}
			return
		default:
		}

		// Perform channel search
		channels, err := app.Channels.All()
		if err != nil {
			errorsCh <- searchError{Source: "channels", Err: err}
			return
		}
		channelsCh <- channels
	}()

	// TODO(human): Exercise 4 Part 4 - Wait for all goroutines and collect results
	// Close error channel when all workers are done
	go func() {
		wg.Wait()
		close(errorsCh)
	}()

	// Collect results
	result := &SearchResult{
		Users:    make([]*models.User, 0),
		Posts:    make([]*models.Post, 0),
		Channels: make([]*models.Channel, 0),
		Errors:   make([]error, 0),
	}

	// Receive from each result channel exactly once
	for range 3 {
		select {
		case users := <-usersCh:
			result.Users = users
		case posts := <-postsCh:
			result.Posts = posts
		case channels := <-channelsCh:
			result.Channels = channels
		}
	}

	// Collect errors
	for err := range errorsCh {
		result.Errors = append(result.Errors, err)
	}

	result.Duration = time.Since(start)

	// Return error if any search failed
	if len(result.Errors) > 0 {
		return result, fmt.Errorf("search completed with %d errors", len(result.Errors))
	}

	return result, nil
}

// enrichPostsWithChannels adds channel information to posts
// This should run AFTER concurrent search completes
func enrichPostsWithChannels(app *app.App, posts []*models.Post, channels []*models.Channel) []*models.Post {
	// Create channel lookup map for O(1) access instead of O(n²) nested loops
	channelMap := make(map[int64]*models.Channel)
	for _, ch := range channels {
		channelMap[ch.ID] = ch
	}

	// Enrich posts with channel data
	for i := range posts {
		// Get channel IDs for this post
		channelIDs, err := app.Channels.GetChannelIDFromPost(posts[i].ID)
		if err != nil || len(channelIDs) == 0 {
			continue
		}

		posts[i].ChannelID = channelIDs[0]

		// Lookup channel name from map
		if channel, ok := channelMap[posts[i].ChannelID]; ok {
			posts[i].ChannelName = channel.Name
		}
	}

	return posts
}
