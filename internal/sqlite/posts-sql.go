package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type PostModel struct {
	DB *sql.DB
}

// Insert a new post into the database
func (m *PostModel) Insert(ctx context.Context, title, content, images, author, authorAvatar string, authorID models.UUIDField, commentable, isFlagged bool) (int64, error) {
	stmt := "INSERT INTO Posts (Title, Content, Images, Created, Author, AuthorAvatar, AuthorID, IsCommentable, IsFlagged) VALUES (?, ?, ?, DateTime('now'), ?, ?, ?, ?, ?)"
	result, err := m.DB.ExecContext(ctx, stmt, title, content, images, author, authorAvatar, authorID, commentable, isFlagged)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	// fmt.Printf(ErrorMsgs.KeyValuePair, "Inserting a new post with ID: ", id)

	return int64(id), nil
}

func (m *PostModel) All(ctx context.Context) ([]*models.Post, error) {
	stmt := "SELECT * FROM Posts ORDER BY Created DESC"
	rows, selectErr := m.DB.QueryContext(ctx, stmt)
	if selectErr != nil {
		return nil, fmt.Errorf("failed to query all posts: %w", selectErr)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()

	var Posts []*models.Post
	for rows.Next() {
		p := models.Post{}
		scanErr := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.Images,
			&p.Created,
			&p.Updated,
			&p.IsCommentable,
			&p.Author,
			&p.AuthorID,
			&p.AuthorAvatar,
			&p.IsFlagged)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", scanErr)
		}
		Posts = append(Posts, &p)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating post rows: %w", rowsErr)
	}

	return Posts, nil
}

func (m *PostModel) GetPostsByUserID(ctx context.Context, user models.UUIDField) ([]*models.Post, error) {
	stmt := "SELECT * FROM posts WHERE AuthorID = ? ORDER BY ID DESC"
	rows, err := m.DB.QueryContext(ctx, stmt, user)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts by user ID: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()

	var Posts []*models.Post
	for rows.Next() {
		p := models.Post{}
		scanErr := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.Images,
			&p.Created,
			&p.Updated,
			&p.IsCommentable,
			&p.Author,
			&p.AuthorID,
			&p.AuthorAvatar,
			&p.IsFlagged)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", scanErr)
		}
		Posts = append(Posts, &p)
	}
	return Posts, nil
}

func (m *PostModel) GetPostsByChannel(ctx context.Context, channel int64) ([]*models.Post, error) {
	stmt := "SELECT * FROM Posts WHERE ID IN (SELECT PostID FROM PostChannels WHERE ChannelID = ?) ORDER BY Created DESC"
	rows, err := m.DB.QueryContext(ctx, stmt, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts by channel: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()

	var Posts []*models.Post
	for rows.Next() {
		p := models.Post{}
		scanErr := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.Images,
			&p.Created,
			&p.Updated,
			&p.IsCommentable,
			&p.Author,
			&p.AuthorID,
			&p.AuthorAvatar,
			&p.IsFlagged)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", scanErr)
		}
		Posts = append(Posts, &p)
	}

	return Posts, nil
}

func (m *PostModel) GetPostByID(ctx context.Context, id int64) (models.Post, error) {
	stmt := "SELECT * FROM Posts WHERE ID = ?"
	row := m.DB.QueryRowContext(ctx, stmt, id)
	p := models.Post{}
	err := row.Scan(
		&p.ID,
		&p.Title,
		&p.Content,
		&p.Images,
		&p.Created,
		&p.Updated,
		&p.IsCommentable,
		&p.Author,
		&p.AuthorID,
		&p.AuthorAvatar,
		&p.IsFlagged)
	if err != nil {
		return p, fmt.Errorf("failed to get post by ID %d: %w", id, err)
	}

	return p, nil
}

func (m *PostModel) GetAllChannelPostsForUser(ctx context.Context, ID models.UUIDField) ([]models.Post, error) {
	stmt := "SELECT * From posts WHERE ID IN (SELECT ChannelID FROM Memberships WHERE UserID = ?)"
	rows, err := m.DB.QueryContext(ctx, stmt, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse results
	posts := make([]models.Post, 0) // Pre-allocate slice
	for rows.Next() {
		c, err := parsePostRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error parsing row: %w", err)
		}
		posts = append(posts, *c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return posts, nil
}

// FindCurrentPost queries the database for any post column that contains the values and returns that post
func (m *PostModel) FindCurrentPost(ctx context.Context, column string, value any) ([]models.Post, error) {
	// Validate column name to prevent SQL injection
	validColumns := map[string]bool{
		"id":            true,
		"title":         true,
		"content":       true,
		"images":        true,
		"created":       true,
		"updated":       true,
		"isCommentable": true,
		"author":        true,
		"authorID":      true,
		"authorAvatar":  true,
		"isFlagged":     true,
	}

	if !validColumns[column] {
		return nil, fmt.Errorf("invalid column name: %s", column)
	}

	// Base query
	query := fmt.Sprintf("SELECT id, title, content, images, created, isCommentable, author, authorID, authorAvatar,  isFlagged FROM posts WHERE %s = ? LIMIT 1", column)

	row := m.DB.QueryRowContext(ctx, query, value)

	// Parse result into a single post
	var posts []models.Post
	var post models.Post
	var avatar, images sql.NullString

	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&images,
		&post.Created,
		&post.Updated,
		&post.IsCommentable,
		&post.Author,
		&post.AuthorID,
		&avatar,
		&post.IsFlagged,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No post found
		}
		return nil, fmt.Errorf("failed to find post by column %s: %w", column, err)
	}

	post.AuthorAvatar = avatar.String
	post.Images = images.String
	posts = append(posts, post)

	return posts, nil
}

func parsePostRows(rows *sql.Rows) (*models.Post, error) {
	var post models.Post

	if err := rows.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Images,
		&post.Created,
		&post.Updated,
		&post.IsCommentable,
		&post.Author,
		&post.AuthorID,
		&post.IsFlagged,
		&post.ChannelID,
		&post.ChannelName,
		&post.Likes,
		&post.Dislikes,
		&post.CommentsCount,
		&post.AuthorAvatar,
		&post.Comments,
	); err != nil {
		return nil, err
	}

	models.UpdateTimeSince(&post)
	return &post, nil
}
