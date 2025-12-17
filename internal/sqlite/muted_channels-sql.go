package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type MutedChannelModel struct {
	DB *sql.DB
}

func (m *MutedChannelModel) Insert(ctx context.Context, authorID, postID int) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for Insert in MutedChannels: %w", err)
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

	stmt := "INSERT INTO MutedChannels (UserID, ChannelID, Created) VALUES (?, ?, DateTime('now'))"
	_, err = tx.Exec(stmt, authorID, postID)
	if err != nil {
		return fmt.Errorf("failed to execute statement for Insert in MutedChannels: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for Insert in MutedChannels: %w", err)
	}

	return nil
}

func (m *MutedChannelModel) All(ctx context.Context) ([]models.MutedChannel, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for All in MutedChannels: %w", err)
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

	stmt := "SELECT ID, UserID, ChannelID, Created FROM MutedChannels ORDER BY ID DESC"
	rows, err := tx.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	var MutedChannels []models.MutedChannel
	for rows.Next() {
		p := models.MutedChannel{}
		err = rows.Scan(&p.ID, &p.UserID, &p.ChannelID)
		if err != nil {
			return nil, err
		}
		MutedChannels = append(MutedChannels, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction for All in MutedChannels: %w", err)
	}

	return MutedChannels, nil
}
