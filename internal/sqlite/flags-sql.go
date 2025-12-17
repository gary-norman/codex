package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type FlagModel struct {
	DB *sql.DB
}

func (m *FlagModel) Insert(ctx context.Context, flagType, content string, approved bool, authorID, channelID, flaggedUserID, flaggedPostID, flaggedCommentID int) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for Insert in Flags: %w", err)
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

	stmt := "INSERT INTO Flags (Flag_type, Content, Created, Approved, AuthorID, ChannelID, Flagged_userID, Flagged_postID, Flagged_commentID) VALUES (?, ?, DateTime('now'), ?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(stmt, flagType, content, approved, authorID, channelID, flaggedUserID, flaggedPostID, flaggedCommentID)
	if err != nil {
		return fmt.Errorf("failed to execute statement for Insert in Flags: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for Insert in Flags: %w", err)
	}

	return nil
}

func (m *FlagModel) All(ctx context.Context) ([]models.Flag, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for All in Flags: %w", err)
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

	stmt := "SELECT ID, Flag_type, Content, Created, Approved, AuthorID, ChannelID, Flagged_userID, Flagged_postID, Flagged_commentID FROM Flags ORDER BY ID DESC"
	rows, err := tx.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	var Flags []models.Flag
	for rows.Next() {
		p := models.Flag{}
		err = rows.Scan(&p.ID, &p.FlagType, &p.Content, &p.Created, &p.Approved, &p.AuthorID, &p.ChannelID, &p.FlaggedUserID, &p.FlaggedPostID, &p.FlaggedCommentID)
		if err != nil {
			return nil, err
		}
		Flags = append(Flags, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction for All in Flags: %w", err)
	}

	return Flags, nil
}
