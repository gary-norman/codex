package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type ModModel struct {
	DB *sql.DB
}

type ModIds struct {
	UserID    models.UUIDField `json:"userId"`
	ChannelID int64            `json:"channelId"`
}

func (m *ModModel) All() ([]models.Mod, error) {
	stmt := "SELECT * FROM Mods ORDER BY ID DESC"
	rows, queryErr := m.DB.Query(stmt)
	if queryErr != nil {
		return nil, fmt.Errorf("failed to query all mods: %w", queryErr)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()

	mods := make([]models.Mod, 0)
	for rows.Next() {
		m, err := parseModeratorRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error parsing row: %w", err)
		}
		mods = append(mods, *m)
	}
	return mods, nil
}

func (m *ModModel) AddModeration(userID models.UUIDField, channelID int64) error {
	stmt := "INSERT INTO Mods (UserID, ChannelID, Created) VALUES (?, ?, DateTime('now'))"
	_, err := m.DB.Exec(stmt, userID, channelID)
	return err
}

func (m *ModModel) GetModdedChannelsForUser(models.UUIDField) ([]models.Mod, error) {
	stmt := "SELECT * From Channels WHERE ID IN (SELECT ChannelID FROM Mods WHERE UserID = ?)"
	rows, queryErr := m.DB.Query(stmt)
	if queryErr != nil {
		return nil, fmt.Errorf("failed to query modded channels for user: %w", queryErr)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()

	mods := make([]models.Mod, 0)
	for rows.Next() {
		m, err := parseModeratorRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error parsing row: %w", err)
		}
		mods = append(mods, *m)
	}
	return mods, nil
}

func (m *ModModel) GetModdedChannelID(ID models.UUIDField) ([]int64, error) {
	stmt := ("SELECT ChannelID FROM Mods WHERE UserID = ?")

	rows, queryErr := m.DB.Query(stmt, ID)
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()

	var ids []int64
	for rows.Next() {
		var m int64
		if err := rows.Scan(&m); err != nil {
			return nil, err
		}
		ids = append(ids, m)
	}
	return ids, nil
}

func (m *ModModel) GetModerator(ID int64) ([]models.UUIDField, error) {
	stmt := ("SELECT UserID FROM Mods WHERE ChannelID = ?")

	rows, queryErr := m.DB.Query(stmt, ID)
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()

	var ids []models.UUIDField
	for rows.Next() {
		var m models.UUIDField
		if err := rows.Scan(&m); err != nil {
			return nil, err
		}
		ids = append(ids, m)
	}
	return ids, nil
}

func parseModeratorRow(row *sql.Row) (*models.Mod, error) {
	var mod models.Mod

	if err := row.Scan(
		&mod.ID,
		&mod.UserID,
		&mod.ChannelID,
		&mod.Created,
	); err != nil {
		return nil, fmt.Errorf("failed to scan moderator row: %w", err)
	}
	models.UpdateTimeSince(&mod)
	return &mod, nil
}

func parseModeratorRows(rows *sql.Rows) (*models.Mod, error) {
	var mod models.Mod

	if err := rows.Scan(
		&mod.ID,
		&mod.UserID,
		&mod.ChannelID,
		&mod.Created,
	); err != nil {
		return nil, fmt.Errorf("failed to scan moderator row: %w", err)
	}
	models.UpdateTimeSince(&mod)
	return &mod, nil
}
