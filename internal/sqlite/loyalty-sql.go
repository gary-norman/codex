package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type LoyaltyModel struct {
	DB *sql.DB
}

func (m *LoyaltyModel) InsertLoyalty(ctx context.Context, follower, following models.UUIDField) error {
	err := m.InsertFollowing(ctx, follower, following)
	if err != nil {
		fmt.Println("Error adding a following")
		return errors.New(err.Error())
	}

	err = m.InsertFollower(ctx, following, follower)
	if err != nil {
		fmt.Println("Error adding a follower")
		return errors.New(err.Error())
	}

	return err
}

// InsertFollower inserts a
func (m *LoyaltyModel) InsertFollower(ctx context.Context, user, follower models.UUIDField) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	// fmt.Println("Beginning UPDATE transaction")
	if err != nil {
		return fmt.Errorf("failed to begin transaction for Insert Follower: %w", err)
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

	query := "INSERT INTO Followers (UserID, FollowerUserID) VALUES (?, ?)"
	_, InsertErr := tx.ExecContext(ctx, query, user, follower)
	// fmt.Printf("Updating Comments, where reactionID: %v, PostID: %v and UserID: %v with Liked: %v, Disliked: %v\n", reactionID, reactedPostID, authorID, liked, disliked)
	if InsertErr != nil {
		return fmt.Errorf("failed to execute Insert query in Insert Follower: %w", err)
	}

	// Commit the transaction
	commitErr := tx.Commit()
	// fmt.Println("Committing UPDATE transaction")
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction for Insert query in Insert Follower: %w", err)
	}

	return commitErr
}

func (m *LoyaltyModel) CountUsers(ctx context.Context, userID models.UUIDField) (followers, following int, err error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	// fmt.Println("Beginning DELETE transaction")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction for CountUsers: %w", err)
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

	query1 := `SELECT COUNT(*) AS FollowingCount
             FROM Following
             WHERE UserID = ?`

	query2 := `SELECT COUNT(*) AS FollowersCount
             FROM Followers
             WHERE UserID = ?`

	var followingCount, followersCount sql.NullInt64

	// Run the query
	err = tx.QueryRowContext(ctx, query1, userID).Scan(&followingCount)
	if err != nil {
		return 0, 0, err
	}

	// Run the query
	err = tx.QueryRowContext(ctx, query2, userID).Scan(&followersCount)
	if err != nil {
		return 0, 0, err
	}

	// Commit the transaction
	commitErr := tx.Commit()
	// fmt.Println("Committing UPDATE transaction")
	if commitErr != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction for CountUsers: %w", err)
	}

	followers = int(followersCount.Int64)
	following = int(followingCount.Int64)

	return followers, following, err
}

// Delete removes an entry in the Following table by ID
func (m *LoyaltyModel) Delete(ctx context.Context, followingID, followersID models.UUIDField) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	// fmt.Println("Beginning DELETE transaction")
	if err != nil {
		return fmt.Errorf("failed to begin transaction for Delete in Following: %w", err)
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

	query1 := `DELETE FROM Following WHERE ID = ?`
	// Execute the query, dereferencing the pointers for ID values
	_, err = tx.ExecContext(ctx, query1, followingID)
	// fmt.Printf("Deleting from Reactions where commentID: %v\n", commentID)
	if err != nil {
		return fmt.Errorf("failed to execute Delete query: %w", err)
	}

	query2 := `DELETE FROM Followers WHERE ID = ?`
	// Execute the query, dereferencing the pointers for ID values
	_, err = tx.ExecContext(ctx, query2, followersID)
	// fmt.Printf("Deleting from Reactions where commentID: %v\n", commentID)
	if err != nil {
		return fmt.Errorf("failed to execute Delete query: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	// fmt.Println("Committing DELETE transaction")
	if err != nil {
		return fmt.Errorf("failed to commit transaction for Delete in Following: %w", err)
	}

	return err
}

// InsertFollowing inserts a new user to the Following list of a target use
func (m *LoyaltyModel) InsertFollowing(ctx context.Context, user, following models.UUIDField) error {
	// Begin the transaction
	tx, err := m.DB.Begin()
	// fmt.Println("Beginning UPDATE transaction")
	if err != nil {
		return fmt.Errorf("failed to begin transaction for Insert in Following: %w", err)
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

	query := "INSERT INTO Following (UserID, FollowingUserID) VALUES (?, ?)"
	_, InsertErr := tx.ExecContext(ctx, query, user, following)
	if InsertErr != nil {
		return fmt.Errorf("failed to execute Insert query in Insert Following: %w", err)
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction in Insert Following: %w", err)
	}

	return commitErr
}

func (m *LoyaltyModel) All(ctx context.Context) ([]models.Loyalty, error) {
	// Begin the transaction
	tx, err := m.DB.Begin()
	// fmt.Println("Beginning UPDATE transaction")
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for Loyalty -> All: %w", err)
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

	query := "SELECT ID, Follower, Followee FROM Loyalty ORDER BY ID DESC"
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var Loyalty []models.Loyalty
	for rows.Next() {
		p := models.Loyalty{}
		err = rows.Scan(&p.ID, &p.Follower, &p.Followee)
		if err != nil {
			return nil, err
		}
		Loyalty = append(Loyalty, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, fmt.Errorf("failed to commit transaction in Loyalty -> All: %w", err)
	}

	return Loyalty, nil
}
