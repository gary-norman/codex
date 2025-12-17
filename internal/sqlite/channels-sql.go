package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/gary-norman/forum/internal/models"
)

type ChannelModel struct {
	DB *sql.DB
}

// RandomInt Function to get a random integer between 0 and the max number, for go templates
func RandomInt(max int) int {
	return rand.Intn(max)
}

func (m *ChannelModel) Insert(ctx context.Context, ownerID models.UUIDField, name, description, avatar, banner string, privacy, isFlagged, isMuted bool) error {
	stmt := "INSERT INTO Channels (OwnerID, Name, Description, Created, Avatar, Banner, Privacy, IsFlagged, IsMuted) VALUES (?, ?, ?, DateTime('now'), ?, ?, ?, ?, ?)"
	_, err := m.DB.ExecContext(ctx, stmt, ownerID, name, description, avatar, banner, privacy, isFlagged, isMuted)
	return err
}

func (m *ChannelModel) OwnedOrJoinedByCurrentUser(ctx context.Context, ID models.UUIDField) ([]*models.Channel, error) {
	stmt := `
	SELECT c.*,
	COUNT(m.UserID) AS MemberCount
	From Channels c
	LEFT JOIN Memberships m ON c.ID = m.ChannelID
	WHERE c.ID IN (
		SELECT ChannelID FROM Memberships WHERE UserID = ?
	)
	OR c.OwnerID = ?
	GROUP BY c.ID
	ORDER BY Name DESC
	`
	rows, err := m.DB.QueryContext(ctx, stmt, ID, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse results
	channels := make([]*models.Channel, 0) // Pre-allocate slice
	for rows.Next() {
		c, err := parseChannelRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error parsing row: %w", err)
		}
		// FIXME: This is a temporary fix to set the channel as joined:we need to come up with a more robust solution
		c.Joined = true
		// TODO (realtime) get this data from websockets
		rnd := RandomInt(1800)
		c.MembersOnline = rnd
		channels = append(channels, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return channels, nil
}

func (m *ChannelModel) IsUserMemberOfChannel(ctx context.Context, userID models.UUIDField, channelID int64) (bool, error) {
	var exists int
	stmt := `
		SELECT EXISTS (
			SELECT 1 FROM Memberships
			WHERE UserID = ? AND ChannelID = ?
		)
	`
	err := m.DB.QueryRowContext(ctx, stmt, userID, channelID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (m *ChannelModel) GetChannelsByID(ctx context.Context, id int64) ([]*models.Channel, error) {
	stmt := `
	SELECT c.*,
  COUNT(m.UserID) AS MemberCount
	FROM Channels c
	LEFT JOIN Memberships m ON c.ID = m.ChannelID
	WHERE c.ID = ?
	GROUP BY c.ID;
	`
	rows, err := m.DB.QueryContext(ctx, stmt, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse results
	channels := make([]*models.Channel, 0) // Pre-allocate slice
	for rows.Next() {
		c, err := parseChannelRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error parsing row: %w", err)
		}
		// TODO (realtime) get this data from websockets
		rnd := RandomInt(1800)
		c.MembersOnline = rnd
		channels = append(channels, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return channels, nil
}

func (m *ChannelModel) GetChannelByID(ctx context.Context, id int64) (*models.Channel, error) {
	stmt := `
	SELECT c.*,
  COUNT(m.UserID) AS MemberCount
	FROM Channels c
	LEFT JOIN Memberships m ON c.ID = m.ChannelID
	WHERE c.ID = ?
	GROUP BY c.ID;
	`
	rows, err := m.DB.QueryContext(ctx, stmt, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channel models.Channel             // Pre-allocate slice
	channels := make([]*models.Channel, 0) // Pre-allocate slice
	for rows.Next() {
		c, err := parseChannelRows(rows)
		if err != nil {
			return nil, err
		}
		// TODO (realtime) get this data from websockets
		rnd := RandomInt(1800)
		c.MembersOnline = rnd
		channels = append(channels, c)
	}
	if len(channels) == 0 {
		return &channel, fmt.Errorf("no channel found for ID %d", id)
	}
	return channels[0], nil
}

func (m *ChannelModel) GetNameOfChannel(ctx context.Context, channelID int64) (string, error) {
	stmt := "SELECT Name FROM Channels WHERE ID = ?)"
	rows, err := m.DB.QueryContext(ctx, stmt, channelID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var username string
	for rows.Next() {
		if err := rows.Scan(&username); err != nil {
			return "", err
		}
	}
	return username, nil
}

func (m *ChannelModel) GetNameOfChannelOwner(ctx context.Context, channelID int64) (string, error) {
	stmt := "SELECT Username FROM Users WHERE ID = (SELECT OwnerID FROM Channels WHERE ID = ?)"
	rows, err := m.DB.QueryContext(ctx, stmt, channelID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var username string
	for rows.Next() {
		if err := rows.Scan(&username); err != nil {
			return "", err
		}
	}
	return username, nil
}

func (m *ChannelModel) All(ctx context.Context) ([]*models.Channel, error) {
	stmt := `
-- 	SELECT c.*,
SELECT c.ID, c.OwnerID, c.Name, c.Avatar, c.Banner, c.Description, c.Created, c.Updated, c.Privacy, c.IsMuted,  c.IsFlagged,
  COUNT(m.UserID) AS MemberCount
	FROM Channels c
	LEFT JOIN Memberships m ON c.ID = m.ChannelID
	GROUP BY c.ID;
	`
	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows in ChannelModel.All: %v", closeErr)
		}
	}()

	channels := make([]*models.Channel, 0) // Pre-allocate slice
	for rows.Next() {
		c, err := parseChannelRows(rows)
		if err != nil {
			return nil, err
		}
		// TODO (realtime) get this data freom websockets
		rnd := RandomInt(1800)
		c.MembersOnline = rnd
		channels = append(channels, c)
	}
	// fmt.Printf(ErrorMsgs.KeyValuePair, "Total channels", len(Channels))
	return channels, nil
}

func isValidColumn(column string) bool {
	validColumns := map[string]bool{
		"ID":          true,
		"OwnerID":     true,
		"Name":        true,
		"Avatar":      true,
		"Banner":      true,
		"Description": true,
		"Created":     true,
		"Privacy":     true,
		"IsMuted":     true,
		"IsFlagged":   true,
	}
	return validColumns[column]
}

func (m *ChannelModel) AddPostToChannel(ctx context.Context, channelID, postID int64) error {
	stmt := "INSERT INTO PostChannels (ChannelID, PostID, Created) VALUES (?, ?, DateTime('now'))"
	_, err := m.DB.ExecContext(ctx, stmt, channelID, postID)
	if err != nil {
		return fmt.Errorf("failed to add post %d to channel %d: %w", postID, channelID, err)
	}
	return nil
}

func (m *ChannelModel) GetPostIDsFromChannel(ctx context.Context, channelID int64) ([]int64, error) {
	var postIDs []int64
	stmt := "SELECT PostID FROM PostChannels WHERE ChannelID = ?"
	rows, err := m.DB.QueryContext(ctx, stmt, channelID)
	if err != nil {
		return postIDs, fmt.Errorf("failed to get post IDs from channel %d: %w", channelID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var postID int64
		if err := rows.Scan(&postID); err != nil {
			return postIDs, fmt.Errorf("failed to scan post ID from channel %d: %w", channelID, err)
		}
		postIDs = append(postIDs, postID)
	}

	return postIDs, nil
}

func (m *ChannelModel) GetChannelIDFromPost(ctx context.Context, postID int64) ([]int64, error) {
	var channelIDs []int64
	stmt := "SELECT ChannelID FROM PostChannels WHERE PostID = ?"
	rows, err := m.DB.QueryContext(ctx, stmt, postID)
	if err != nil {
		return channelIDs, fmt.Errorf("failed to get channel ID from post %d: %w", postID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var channelID int64
		if err := rows.Scan(&channelID); err != nil {
			return channelIDs, fmt.Errorf("failed to scan channel ID from post %d: %w", postID, err)
		}
		channelIDs = append(channelIDs, channelID)
	}

	if len(channelIDs) == 0 {
		return channelIDs, fmt.Errorf("no channel found for post %d", postID)
	}
	return channelIDs, nil
}

func (m *ChannelModel) GetChannelNameFromID(ctx context.Context, id int64) (string, error) {
	var name string
	stmt := "SELECT Name FROM Channels WHERE ID = ?"
	row := m.DB.QueryRowContext(ctx, stmt, id)
	if err := row.Scan(&name); err != nil {
		return "", fmt.Errorf("failed to get channel name for ID %d: %w", id, err)
	}

	return name, nil
}

func parseChannelRow(row *sql.Row) (*models.Channel, error) {
	var channel models.Channel
	var avatar, banner sql.NullString

	if err := row.Scan(
		&channel.ID,
		&channel.OwnerID,
		&channel.Name,
		&avatar,
		&banner,
		&channel.Description,
		&channel.Created,
		&channel.Updated,
		&channel.Privacy,
		&channel.IsMuted,
		&channel.IsFlagged,
		&channel.Members,
	); err != nil {
		return nil, fmt.Errorf("failed to scan channel row: %w", err)
	}

	channel.Avatar = avatar.String
	channel.Banner = banner.String
	models.UpdateTimeSince(&channel)
	return &channel, nil
}

func parseChannelRows(rows *sql.Rows) (*models.Channel, error) {
	var channel models.Channel
	var avatar, banner sql.NullString

	if err := rows.Scan(
		&channel.ID,
		&channel.OwnerID,
		&channel.Name,
		&avatar,
		&banner,
		&channel.Description,
		&channel.Created,
		&channel.Updated,
		&channel.Privacy,
		&channel.IsMuted,
		&channel.IsFlagged,
		&channel.Members,
	); err != nil {
		return nil, fmt.Errorf("failed to scan channel row: %w", err)
	}

	channel.Avatar = avatar.String
	channel.Banner = banner.String
	models.UpdateTimeSince(&channel)
	return &channel, nil
}
