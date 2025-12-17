package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type SavedModel struct {
	DB *sql.DB
}

func (m *SavedModel) Insert(ctx context.Context, postID, commentID, channelID int64) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for Insert in SavedModel: %w", err)
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

	stmt := "INSERT INTO Bookmarks (PostID, CommentID, ChannelID, Created) VALUES (?, ?, ?, DateTime('now'))"
	if _, err = tx.ExecContext(ctx, stmt, postID, commentID, channelID); err != nil {
		return fmt.Errorf("failed to execute statement for Insert in SavedModel: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for Insert in SavedModel: %w", err)
	}

	return nil
}

func (m *SavedModel) All(ctx context.Context) ([]models.Bookmark, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for All in SavedModel: %w", err)
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

	stmt := "SELECT ID, PostID, CommentID, ChannelID, Created FROM Bookmarks ORDER BY ID DESC"
	rows, err := tx.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	var Bookmarks []models.Bookmark
	for rows.Next() {
		p := models.Bookmark{}
		err = rows.Scan(&p.ID, &p.PostID, &p.CommentID, &p.ChannelID, &p.Created)
		if err != nil {
			return nil, err
		}
		Bookmarks = append(Bookmarks, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction for All in SavedModel: %w", err)
	}

	return Bookmarks, nil
}
