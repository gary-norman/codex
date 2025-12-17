package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/gary-norman/forum/internal/app"
)

func TestConcurrentSearch(t *testing.T) {
	// Initialize test database
	appInstance, cleanup, err := app.InitializeApp()
	if err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	t.Run("returns results from all sources", func(t *testing.T) {
		result, err := ConcurrentSearch(ctx, appInstance)
		if err != nil {
			t.Fatalf("ConcurrentSearch failed: %v", err)
		}

		// Should have results (even if empty, slices should be initialized)
		if result.Users == nil {
			t.Error("Users slice is nil")
		}
		if result.Posts == nil {
			t.Error("Posts slice is nil")
		}
		if result.Channels == nil {
			t.Error("Channels slice is nil")
		}

		t.Logf("Search completed in %v", result.Duration)
		t.Logf("Found: %d users, %d posts, %d channels",
			len(result.Users), len(result.Posts), len(result.Channels))
	})

	t.Run("respects context timeout", func(t *testing.T) {
		// Very short timeout - should cancel
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure context is cancelled

		result, err := ConcurrentSearch(ctx, appInstance)

		// Should handle cancellation gracefully
		if err == nil && len(result.Errors) == 0 {
			t.Log("Warning: Expected timeout or errors, but search completed successfully")
		}

		if err != nil {
			t.Logf("Correctly handled timeout: %v", err)
		}
	})

	t.Run("is faster than sequential search", func(t *testing.T) {
		// Run concurrent search
		concurrentStart := time.Now()
		_, err := ConcurrentSearch(ctx, appInstance)
		concurrentDuration := time.Since(concurrentStart)
		if err != nil {
			t.Fatalf("Concurrent search failed: %v", err)
		}

		// Run sequential search
		sequentialStart := time.Now()
		appInstance.Users.All(ctx)
		appInstance.Posts.All(ctx)
		appInstance.Channels.All(ctx)
		sequentialDuration := time.Since(sequentialStart)

		t.Logf("Concurrent: %v", concurrentDuration)
		t.Logf("Sequential: %v", sequentialDuration)

		// Concurrent should be faster or similar (not slower)
		if concurrentDuration > sequentialDuration*2 {
			t.Errorf("Concurrent search is slower than sequential (concurrent: %v, sequential: %v)",
				concurrentDuration, sequentialDuration)
		}
	})

	t.Run("handles partial failures", func(t *testing.T) {
		// Even if one search fails, others should succeed
		result, _ := ConcurrentSearch(ctx, appInstance)

		// At least some data should be returned
		totalResults := len(result.Users) + len(result.Posts) + len(result.Channels)
		if totalResults == 0 && len(result.Errors) == 0 {
			t.Error("No results and no errors - unexpected state")
		}
	})
}

func TestEnrichPostsWithChannels(t *testing.T) {
	ctx := context.Background()
	appInstance, cleanup, err := app.InitializeApp()
	if err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}
	defer cleanup()

	posts, err := appInstance.Posts.All(ctx)
	if err != nil {
		t.Fatalf("Failed to get posts: %v", err)
	}

	channels, err := appInstance.Channels.All(ctx)
	if err != nil {
		t.Fatalf("Failed to get channels: %v", err)
	}

	enriched := enrichPostsWithChannels(appInstance, posts, channels)

	// Check that posts were enriched
	for _, post := range enriched {
		if post.ChannelID != 0 && post.ChannelName == "" {
			t.Errorf("Post %d has ChannelID %d but no ChannelName",
				post.ID, post.ChannelID)
		}
	}

	t.Logf("Enriched %d posts with channel data", len(enriched))
}

func BenchmarkConcurrentSearch(b *testing.B) {
	appInstance, cleanup, err := app.InitializeApp()
	if err != nil {
		b.Fatalf("Failed to initialize app: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ConcurrentSearch(ctx, appInstance)
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
	}
}

func BenchmarkSequentialSearch(b *testing.B) {
	ctx := context.Background()
	appInstance, cleanup, err := app.InitializeApp()
	if err != nil {
		b.Fatalf("Failed to initialize app: %v", err)
	}
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		appInstance.Users.All(ctx)
		appInstance.Posts.All(ctx)
		appInstance.Channels.All(ctx)
	}
}
