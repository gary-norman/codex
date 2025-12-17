package sqlite

import (
	"context"
	"database/sql"

	"github.com/gary-norman/forum/internal/models"
)

type MembershipModel struct {
	DB *sql.DB
}

func (m *MembershipModel) Insert(ctx context.Context, userID models.UUIDField, channelID int64) error {
	query := "INSERT INTO Memberships (UserID, ChannelID, Created) VALUES (?, ?, DateTime('now'))"
	_, err := m.DB.ExecContext(ctx, query, userID, channelID)
	return err
}

func (m *MembershipModel) UserMemberships(ctx context.Context, userID models.UUIDField) ([]models.Membership, error) {
	// fmt.Printf(ErrorMsgs.KeyValuePair, "Checking memberships for UserID", userID)
	query := "SELECT ID, UserID, ChannelID, Created FROM Memberships WHERE UserID = ?"
	rows, queryErr := m.DB.QueryContext(ctx, query, userID)
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows in UserMemberships: %v", closeErr)
		}
	}()
	var memberships []models.Membership
	for rows.Next() {
		p := models.Membership{}
		scanErr := rows.Scan(&p.ID, &p.UserID, &p.ChannelID, &p.Created)
		if scanErr != nil {
			return nil, scanErr
		}
		memberships = append(memberships, p)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	// fmt.Printf(ErrorMsgs.KeyValuePair, "Channels joined by current user", len(memberships))
	return memberships, nil
}

func (m *MembershipModel) All(ctx context.Context) ([]models.Membership, error) {
	query := "SELECT ID, UserID, ChannelID, Created FROM Memberships ORDER BY ID DESC"
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows in All: %v", closeErr)
		}
	}()

	var Memberships []models.Membership
	for rows.Next() {
		p := models.Membership{}
		err = rows.Scan(&p.ID, &p.UserID, &p.ChannelID, &p.Created)
		if err != nil {
			return nil, err
		}
		Memberships = append(Memberships, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return Memberships, nil
}
