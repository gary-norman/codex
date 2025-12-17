package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type ImageModel struct {
	DB *sql.DB
}

func (m *ImageModel) Insert(ctx context.Context, authorID models.UUIDField, postID int64, path string) (int64, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction in Insert Image: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := "INSERT INTO Images (Created, Updated, AuthorID, PostID, Path) VALUES (DateTime('now'), DateTime('now'), ?, ?, ?)"

	result, err := tx.ExecContext(ctx, query, authorID, postID, path)
	if err != nil {
		return 0, err
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return 0, fmt.Errorf("failed to commit transaction in Insert Image: %w", err)
	}

	// Return the ID of the newly inserted image
	imageID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return imageID, nil
}

func (m *ImageModel) All(ctx context.Context) ([]models.Image, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction in All Images: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := "SELECT ID, Created, Updated, AuthorID, PostID, Path FROM Images ORDER BY ID DESC"
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var Images []models.Image
	for rows.Next() {
		p := models.Image{}
		err = rows.Scan(&p.ID, &p.Created, &p.Updated, &p.AuthorID, &p.PostID, &p.Path)
		if err != nil {
			return nil, err
		}
		Images = append(Images, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, fmt.Errorf("failed to commit transaction in All Images: %w", err)
	}

	return Images, nil
}
