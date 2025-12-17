package examples

import (
	"fmt"
	"net/http"

	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/workers"
)

// Example: How to use structured logging in your handlers

// Example 1: Log when a post is created
func ExamplePostCreation(w http.ResponseWriter, r *http.Request, loggerPool *workers.LoggerPool) {
	// ... your handler logic to create post ...

	user := r.Context().Value("user").(*models.User)

	// Simulated post creation
	var newPost models.Post
	fmt.Printf("Creating post by user %s: %s", user.Username, newPost.Title)
	// newPost = ... create post in database ...
	postID := int64(123)
	channelID := int64(5)
	title := "How to use Go concurrency patterns"
	tags := []string{"golang", "concurrency", "tutorial"}

	// Log the event with searchable context
	loggerPool.LogPostCreated(user.ID, postID, channelID, title, tags)

	// Now you can search:
	// - All posts by this user UUID
	// - All posts in this channel
	// - All posts tagged "golang"
	// - All "create" actions by this user
}

// Example 2: Log when a user views a post
func ExamplePostView(w http.ResponseWriter, r *http.Request, loggerPool *workers.LoggerPool) {
	user := r.Context().Value("user").(*models.User)
	postID := int64(123)
	channelID := int64(5)

	// Log the view event
	loggerPool.LogPostViewed(user.ID, postID, channelID)

	// Analytics queries you can now run:
	// - How many times was this post viewed?
	// - Which posts has this user viewed?
	// - Which users viewed this post?
	// - View count trend over time
}

// Example 3: Log when a comment is created
func ExampleCommentCreation(w http.ResponseWriter, r *http.Request, loggerPool *workers.LoggerPool) {
	user := r.Context().Value("user").(*models.User)

	commentContent := "Great explanation of worker pools!"
	postID := int64(123)
	commentID := int64(456)

	// Log comment with content length (useful metric)
	loggerPool.LogCommentCreated(user.ID, commentID, postID, len(commentContent))

	// You can now find:
	// - All comments by this user
	// - All comments on this post
	// - Average comment length by user
	// - Most active commenters
}

// Example 4: Log when a user joins a channel
func ExampleChannelJoin(w http.ResponseWriter, r *http.Request, loggerPool *workers.LoggerPool) {
	user := r.Context().Value("user").(*models.User)
	channelID := int64(5)
	channelName := "golang-discussions"

	// Log the join event with channel name as searchable keyword
	loggerPool.LogChannelJoined(user.ID, channelID, channelName)

	// Search capabilities:
	// - Which channels has this user joined?
	// - Who joined this channel recently?
	// - Channel growth over time
	// - Search for channels by name
}

// Example 5: Log user login
func ExampleUserLogin(w http.ResponseWriter, r *http.Request, loggerPool *workers.LoggerPool) {
	user := r.Context().Value("user").(*models.User)

	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	// Log login with IP and user agent for security monitoring
	loggerPool.LogUserLogin(user.ID, user.Username, ipAddress, userAgent)

	// Security queries:
	// - Logins from unusual IPs
	// - Multiple logins from same user
	// - Login patterns by time of day
	// - Suspicious user agent strings
}

// Example 6: Log search queries
func ExampleSearchQuery(w http.ResponseWriter, r *http.Request, loggerPool *workers.LoggerPool) {
	user := r.Context().Value("user").(*models.User)

	query := r.URL.Query().Get("q")

	// ... perform search ...
	resultCount := 42 // Number of results found

	// Log the search with query as keyword and result count as metric
	loggerPool.LogSearchQuery(user.ID, query, resultCount)

	// Analytics:
	// - Most popular search terms
	// - Search queries with zero results (improve content)
	// - What users search for most
	// - Search patterns (trending topics)
}

// Example 7: Log reactions
func ExampleReaction(w http.ResponseWriter, r *http.Request, loggerPool *workers.LoggerPool) {
	user := r.Context().Value("user").(*models.User)
	postID := int64(123)
	reactionType := "like" // or "dislike"

	// Log the reaction
	loggerPool.LogReaction(user.ID, postID, reactionType)

	// Insights:
	// - Which posts get most likes?
	// - What does this user typically react to?
	// - Like/dislike ratio by post
	// - User engagement patterns
}

// Example 8: Log errors with context
func ExampleErrorLogging(w http.ResponseWriter, r *http.Request, loggerPool *workers.LoggerPool) {
	user := r.Context().Value("user").(*models.User)

	// ... some operation that fails ...
	err := someOperation()
	if err != nil {
		// Log error with full context
		loggerPool.LogApplicationError(
			err,
			user.ID,
			r.URL.Path,
			"create_post",
			map[string]interface{}{
				"post_id":    123,
				"channel_id": 5,
				"action":     "database_insert",
			},
		)
	}

	// Debugging queries:
	// - All errors by this user
	// - All errors on this operation
	// - Error rate by endpoint
	// - Which posts/channels cause most errors
}

func someOperation() error {
	return nil
}

// SEARCH EXAMPLES - How to query the structured logs

func ExampleSearchQueries(loggerPool *workers.LoggerPool) {
	// These would typically be in your admin dashboard handlers

	// 1. Get all activity by a specific user
	// userActivity, err := loggerPool.loggingModel.GetUserActivity("user-uuid-here", 50)

	// 2. Get all interactions with a specific post
	// postInteractions, err := loggerPool.loggingModel.GetPostInteractions(123)

	// 3. Get all activity in a channel
	// channelActivity, err := loggerPool.loggingModel.GetChannelActivity(5, 100)

	// 4. Search for events containing a keyword
	// results, err := loggerPool.loggingModel.SearchByKeyword("golang", 50)

	// 5. Get most active users in last 7 days
	// activeUsers, err := loggerPool.loggingModel.GetMostActiveUsers(7, 10)

	// 6. Get most popular posts in last 30 days
	// popularPosts, err := loggerPool.loggingModel.GetPopularPosts(30, 10)

	// 7. Get all errors for a specific user
	// userErrors, err := loggerPool.loggingModel.GetUserErrors("user-uuid-here", 20)
}

/*
EXAMPLE SQL QUERIES YOU CAN NOW RUN:

-- Find all posts created by a specific user
SELECT
	Timestamp,
	json_extract(Details, '$.post_id') as post_id,
	json_extract(Details, '$.metadata.title') as title,
	json_extract(Details, '$.keywords') as tags
FROM SystemMetrics
WHERE MetricType = 'user_activity'
	AND MetricName = 'post_created'
	AND json_extract(Details, '$.user_id') = 'user-uuid-here';

-- Find all activity for a specific post
SELECT
	Timestamp,
	json_extract(Details, '$.user_id') as user_id,
	json_extract(Details, '$.action') as action,
	MetricName as event_type
FROM SystemMetrics
WHERE json_extract(Details, '$.post_id') = 123
ORDER BY Timestamp DESC;

-- Find most searched keywords
SELECT
	json_extract(Details, '$.keywords') as keyword,
	COUNT(*) as search_count,
	AVG(MetricValue) as avg_results
FROM SystemMetrics
WHERE MetricName = 'search_query'
	AND Timestamp > datetime('now', '-7 days')
GROUP BY json_extract(Details, '$.keywords')
ORDER BY search_count DESC
LIMIT 10;

-- Find users who viewed a post but didn't react
SELECT DISTINCT json_extract(Details, '$.user_id') as viewer_id
FROM SystemMetrics
WHERE MetricName = 'post_viewed'
	AND json_extract(Details, '$.post_id') = 123
	AND json_extract(Details, '$.user_id') NOT IN (
		SELECT json_extract(Details, '$.user_id')
		FROM SystemMetrics
		WHERE MetricName LIKE 'reaction_%'
			AND json_extract(Details, '$.post_id') = 123
	);

-- Engagement funnel: views -> comments -> reactions
SELECT
	json_extract(Details, '$.post_id') as post_id,
	SUM(CASE WHEN MetricName = 'post_viewed' THEN 1 ELSE 0 END) as views,
	SUM(CASE WHEN MetricName = 'comment_created' THEN 1 ELSE 0 END) as comments,
	SUM(CASE WHEN MetricName LIKE 'reaction_%' THEN 1 ELSE 0 END) as reactions
FROM SystemMetrics
WHERE MetricType = 'user_activity'
	AND Timestamp > datetime('now', '-30 days')
GROUP BY json_extract(Details, '$.post_id')
ORDER BY views DESC
LIMIT 20;
*/
