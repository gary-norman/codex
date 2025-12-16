package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type CommentModel struct {
	DB *sql.DB
}

// Upsert inserts or updates a reaction for a specific combination of AuthorID and the parent fields (ChannelID, ReactedPostID, ReactedCommentID). It uses Exists to determine if the reaction already exists.
func (m *CommentModel) Upsert(comment models.Comment) error {
	// Check if the reaction exists
	exists, err := m.Exists(comment)
	if err != nil {
		return fmt.Errorf("failed to check existence of comment: %w", err)
	}

	if exists {
		// If the reaction exists, update it
		// fmt.Println("Updating a reaction which already exists (reactions.go :53)")
		return m.Update(comment)
	}
	// fmt.Println("Inserting a reaction (reactions.go :56)")

	return m.Insert(comment)
}

func (m *CommentModel) Insert(comment models.Comment) error {
	// Begin the transaction
	tx, err := m.DB.Begin()
	// fmt.Println("Beginning INSERT INTO transaction")
	if err != nil {
		return fmt.Errorf("failed to begin transaction for Insert in Comments: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarn("Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Define the SQL statement
	query := `INSERT INTO Comments 
		(Content, Created, Author, AuthorID, AuthorAvatar, ChannelName, ChannelID, CommentedPostID, 
		CommentedCommentID, IsCommentable, IsFlagged, IsReply)
		VALUES (?, DateTime('now'), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Execute the query, dereferencing the pointers is handled by database/sql
	_, err = tx.Exec(query,
		comment.Content,
		comment.Author,
		comment.AuthorID,
		comment.AuthorAvatar,
		comment.ChannelName,
		comment.ChannelID,
		comment.CommentedPostID,
		comment.CommentedCommentID,
		comment.IsCommentable,
		comment.IsFlagged,
		comment.IsReply,
	)
	// fmt.Printf("Inserting row:\nLiked: %v, Disliked: %v, userID: %v, PostID: %v\n", liked, disliked, authorID, parentPostID)
	if err != nil {
		return fmt.Errorf("failed to execute Insert query: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	// fmt.Println("Committing INSERT INTO transaction")
	if err != nil {
		return fmt.Errorf("failed to commit transaction for Insert in Comments: %w", err)
	}

	return nil
}

func (m *CommentModel) Update(comment models.Comment) error {
	//if !isValidParent(*comment.CommentedPostID, *comment.CommentedCommentID) {
	//	return fmt.Errorf("only one of CommentedPostID, or CommentedCommentID must be non-zero")
	//}

	// Begin the transaction
	tx, err := m.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for Insert in Comments: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarn("Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	// TODO add Updated field, which should be populated on update
	// Define the SQL statement
	query := `UPDATE Comments 
		SET Content = ?, IsCommentable = ?, IsFlagged = ?, Author = ?, AuthorAvatar = ?, ChannelName = ?, ChannelID = ?
		WHERE AuthorID = ? AND (CommentedPostID = ? OR CommentedCommentID = ?)`

	// Execute the query
	_, err = tx.Exec(query,
		comment.Content,
		comment.Author,
		comment.AuthorID,
		comment.AuthorAvatar,
		comment.ChannelName,
		comment.ChannelID,
		comment.CommentedPostID,
		comment.CommentedCommentID,
		comment.IsCommentable,
		comment.IsFlagged,
		comment.IsReply)
	// fmt.Printf("Updating Comments, where reactionID: %v, PostID: %v and UserID: %v with Liked: %v, Disliked: %v\n", reactionID, reactedPostID, authorID, liked, disliked)
	if err != nil {
		return fmt.Errorf("failed to execute Update query: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction for Update in Comments: %w", err)
	}

	return nil
}

// Exists helps avoid creating duplicate comments by determining whether a comment for the specific combination of AuthorID, PostID/CommentID and Content
func (m *CommentModel) Exists(comment models.Comment) (bool, error) {
	// SQL query to check if the comment exists with the provided parameters
	stmt := `SELECT EXISTS(
                SELECT 1 FROM Comments
                WHERE AuthorID = ? AND 
                      CommentedPostID = ? AND 
                      CommentedCommentID = ? AND 
                      Content = ?)`

	var exists bool
	err := m.DB.QueryRow(stmt,
		&comment.AuthorID,
		&comment.CommentedPostID,
		&comment.CommentedCommentID,
		&comment.Content).Scan(&exists)

	return exists, err
}

// Delete removes a comment from the database by ID
func (m *CommentModel) Delete(commentID int64) error {
	// Begin the transaction
	tx, err := m.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for Delete in Comments: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarn("Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `DELETE FROM Comments WHERE ID = ?`
	// Execute the query, dereferencing the pointers for ID values
	_, err = m.DB.Exec(query, commentID)
	// fmt.Printf("Deleting from Reactions where commentID: %v\n", commentID)
	if err != nil {
		return fmt.Errorf("failed to execute Delete query: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction for Delete in Comments: %w", err)
	}

	return nil
}

func (m *CommentModel) GetCommentByPostID(id int64) ([]models.Comment, error) {
	// Begin the transaction
	tx, err := m.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for GetCommentByPostID in Comments: %w", err)
	}

	if m == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	stmt := "SELECT * FROM Comments WHERE CommentedPostID = ? ORDER BY ID DESC"
	rows, err := m.DB.Query(stmt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments by post ID %d: %w", id, err)
	}
	defer func() {
		if p := recover(); p != nil {
			models.LogWarn("Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()
	var comments []models.Comment
	for rows.Next() {
		c := models.Comment{}
		scanErr := rows.Scan(
			&c.ID,
			&c.Content,
			&c.Created,
			&c.Updated,
			&c.CommentedPostID,
			&c.CommentedCommentID,
			&c.IsCommentable,
			&c.IsFlagged,
			&c.IsReply,
			&c.Author,
			&c.AuthorID,
			&c.AuthorAvatar,
			&c.ChannelName,
			&c.ChannelID,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", scanErr)
		}
		comments = append(comments, c)
	}
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction for GetCommentByPostID in Comments: %w", err)
	}

	return comments, nil
}

func (m *CommentModel) GetCommentByCommentID(id int64) ([]models.Comment, error) {
	// Begin the transaction
	tx, err := m.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for GetCommentByCommentID in Comments: %w", err)
	}

	if m == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	stmt := "SELECT * FROM Comments WHERE CommentedCommentID = ? ORDER BY ID DESC"
	rows, err := m.DB.Query(stmt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments by comment ID %d: %w", id, err)
	}
	defer func() {
		if p := recover(); p != nil {
			models.LogWarn("Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()
	var comments []models.Comment
	for rows.Next() {
		c := models.Comment{}
		scanErr := rows.Scan(
			&c.ID,
			&c.Content,
			&c.Created,
			&c.Updated,
			&c.AuthorID,
			&c.ChannelID,
			&c.IsReply,
			&c.CommentedPostID,
			&c.CommentedCommentID,
			&c.IsFlagged,
			&c.Author,
			&c.AuthorAvatar,
			&c.ChannelName,
			&c.IsCommentable,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", scanErr)
		}
		comments = append(comments, c)
	}
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction for GetCommentByCommentID in Comments: %w", err)
	}

	return comments, nil
}

func (m *CommentModel) All() ([]models.Comment, error) {
	// Begin the transaction
	tx, err := m.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for All in Comments: %w", err)
	}

	stmt := "SELECT ID, Content, Created, Author, AuthorID, AuthorAvatar, ChannelName, ChannelID, CommentedPostID, CommentedCommentID, IsCommentable, IsFlagged FROM Comments ORDER BY ID DESC"

	if m == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	if m.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	rows, selectErr := m.DB.Query(stmt)
	if selectErr != nil {
		return nil, fmt.Errorf("failed to query all comments: %w", selectErr)
	}

	defer func() {
		if p := recover(); p != nil {
			models.LogWarn("Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()

	var Comments []models.Comment
	for rows.Next() {
		p := models.Comment{}

		scanErr := rows.Scan(
			&p.ID,
			&p.Content,
			&p.Created,
			&p.Updated,
			&p.Author,
			&p.AuthorID,
			&p.AuthorAvatar,
			&p.ChannelName,
			&p.ChannelID,
			&p.CommentedPostID,
			&p.CommentedCommentID,
			&p.IsFlagged,
			&p.IsCommentable)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", scanErr)
		}
		Comments = append(Comments, p)
	}
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction for All in Comments: %w", err)
	}

	return Comments, nil
}

// GetComment checks if a user has already commented on a post or comment. It retrieves already existing reactions.
func (m *CommentModel) GetComment(authorID int, reactedPostID int, reactedCommentID int64) (*models.Reaction, error) {
	var reaction models.Reaction
	var stmt string

	// Build the SQL query depending on whether the reaction is to a post or comment
	if reactedPostID != 0 {
		stmt = `SELECT ID, Created, AuthorID, CommentedPostID, CommentedCommentID, IsCommentable, IsFlagged 
				FROM Comments 
				WHERE AuthorID = ? AND 
				      CommentedPostID = ?`
	} else if reactedCommentID != 0 {
		stmt = `SELECT ID, Liked, Disliked, AuthorID, Created, ReactedPostID, ReactedCommentID 
				FROM Reactions 
				WHERE AuthorID = ? AND 
				      CommentedCommentID = ?`
	} else {
		return nil, nil
	}

	// Query the database
	row := m.DB.QueryRow(stmt, authorID, reactedPostID)
	if reactedCommentID != 0 {
		row = m.DB.QueryRow(stmt, authorID, reactedCommentID)
	}

	err := row.Scan(&reaction.ID, &reaction.Liked, &reaction.Disliked, &reaction.AuthorID, &reaction.Created, &reaction.ReactedPostID, &reaction.ReactedCommentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No reaction found
			return nil, nil
		}
		// Other errors
		return nil, fmt.Errorf("failed to fetch reaction: %w", err)
	}

	// Return the existing reaction
	return &reaction, nil
}
