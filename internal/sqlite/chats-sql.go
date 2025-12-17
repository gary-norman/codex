package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type ChatModel struct {
	DB *sql.DB
}

func (c *ChatModel) CreateChat(ctx context.Context, chatType, name string, groupID, buddyID models.NullableUUIDField) (models.UUIDField, error) {
	chatID := models.NewUUIDField()
	query := "INSERT INTO Chats (ID, Type, Name, GroupID, BuddyID, Created) VALUES (?, ?, ?, ?, ?, DateTime('now'))"
	_, err := c.DB.ExecContext(ctx, query, chatID, chatType, name, groupID, buddyID)
	if err != nil {
		return models.UUIDField{}, fmt.Errorf("failed to insert chat: %w", err)
	}

	return chatID, nil
}

func (c *ChatModel) CreateChatMessage(ctx context.Context, chatID, userID models.UUIDField, message string) (models.UUIDField, error) {
	messageID := models.NewUUIDField()
	query := "INSERT INTO Messages (ID, ChatID, UserID, Created, Content) VALUES (?, ?, ?, DateTime('now'), ?)"
	_, err := c.DB.ExecContext(ctx, query, messageID, chatID, userID, message)
	if err != nil {
		return models.UUIDField{}, fmt.Errorf("failed to insert message: %w", err)
	}

	return messageID, nil
}

func (c *ChatModel) AttachUserToChat(ctx context.Context, chatID, userID models.UUIDField) error {
	query := "INSERT INTO ChatUsers (ChatID, UserID) VALUES (?, ?)"
	_, err := c.DB.ExecContext(ctx, query, chatID, userID)
	if err != nil {
		return fmt.Errorf("failed to attach user to chat: %w", err)
	}

	return nil
}

func (c *ChatModel) GetUserChatIDs(ctx context.Context, userID models.UUIDField) ([]models.UUIDField, error) {
	query := `SELECT ChatID FROM ChatUsers WHERE UserID = ?`
	rows, err := c.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user chat IDs: %w", err)
	}
	defer rows.Close()

	var chatIDs []models.UUIDField
	for rows.Next() {
		var chatID models.UUIDField
		if err := rows.Scan(&chatID); err != nil {
			return nil, fmt.Errorf("failed to scan chat ID: %w", err)
		}
		chatIDs = append(chatIDs, chatID)
	}

	return chatIDs, nil
}

// GetChat retrieves a single chat by its ID
func (c *ChatModel) GetChat(ctx context.Context, chatID models.UUIDField) (*models.Chat, error) {
	// Begin the transaction
	tx, err := c.DB.BeginTx(ctx, nil)
	// fmt.Println("Beginning UPDATE transaction")
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for GetChat: %w", err)
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

	query := "SELECT ID, Type, Name, Created, LastActive, GroupID, BuddyID FROM Chats WHERE ID = ?"
	row := tx.QueryRowContext(ctx, query, chatID)

	var chat models.Chat
	var buddyID, groupID models.NullableUUIDField

	err = row.Scan(&chat.ID, &chat.ChatType, &chat.Name, &chat.Created, &chat.LastActive, &groupID, &buddyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chat not found: %s", chatID)
		}
		return nil, fmt.Errorf("failed to scan chat: %w", err)
	}

	if groupID.Valid {
		chat.Group.ID = groupID.UUID
	}
	if buddyID.Valid {
		chat.Buddy = &models.User{ID: buddyID.UUID}
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, fmt.Errorf("failed to commit transaction for GetChat: %w", commitErr)
	}

	return &chat, nil
}

// GetUserChats retrieves all chats for a specific user
func (c *ChatModel) GetUserChats(ctx context.Context, userID models.UUIDField) ([]models.Chat, error) {
	query := `
		SELECT c.ID, c.Type, c.Name, c.Created, c.LastActive, c.GroupID, c.BuddyID
		FROM Chats c
		INNER JOIN ChatUsers cu ON c.ID = cu.ChatID
		WHERE cu.UserID = ?
		ORDER BY c.LastActive DESC
	`

	rows, err := c.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user chats: %w", err)
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		var buddyID, groupID models.NullableUUIDField

		err := rows.Scan(&chat.ID, &chat.ChatType, &chat.Name, &chat.Created, &chat.LastActive, &groupID, &buddyID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat: %w", err)
		}

		if groupID.Valid {
			chat.Group.ID = groupID.UUID
		}
		if buddyID.Valid {
			chat.Buddy = &models.User{ID: buddyID.UUID}
		}

		chats = append(chats, chat)
	}

	return chats, nil
}

// GetChatMessages retrieves all messages for a specific chat
func (c *ChatModel) GetChatMessages(ctx context.Context, chatID models.UUIDField) ([]models.ChatMessage, error) {
	// Begin the transaction
	tx, err := c.DB.BeginTx(ctx, nil)
	// fmt.Println("Beginning UPDATE transaction")
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for GetChatMessages: %w", err)
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

	query := `
		SELECT
			m.ID, m.ChatID, m.Created, m.Content,
			u.ID, u.Username, u.EmailAddress, u.Avatar, u.Banner,
			u.Description, u.Usertype, u.Created, u.Updated, u.IsFlagged,
			u.SessionToken, u.CSRFToken, u.HashedPassword
		FROM Messages m
		LEFT JOIN Users u ON m.UserID = u.ID
		WHERE m.ChatID = ?
		ORDER BY m.Created ASC
	`

	rows, err := tx.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chat messages: %w", err)
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var message models.ChatMessage
		var user models.User

		// Use sql.Null types for potentially NULL user fields
		var (
			userID         sql.NullString
			username       sql.NullString
			email          sql.NullString
			avatar         sql.NullString
			banner         sql.NullString
			description    sql.NullString
			usertype       sql.NullString
			userCreated    sql.NullTime
			userUpdated    sql.NullTime
			isFlagged      sql.NullBool
			sessionToken   sql.NullString
			csrfToken      sql.NullString
			hashedPassword sql.NullString
		)

		err := rows.Scan(
			&message.ID, &message.ChatID, &message.Created, &message.Content,
			&userID, &username, &email, &avatar, &banner,
			&description, &usertype, &userCreated, &userUpdated, &isFlagged,
			&sessionToken, &csrfToken, &hashedPassword,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}

		// Only populate Sender if user exists (LEFT JOIN might return NULLs)
		if userID.Valid {
			id, err := models.UUIDFieldFromString(userID.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse user ID: %w", err)
			}

			user.ID = id
			user.Username = username.String
			user.Email = email.String
			user.Avatar = avatar.String
			user.Banner = banner.String
			user.Description = description.String
			user.Usertype = usertype.String
			user.Created = userCreated.Time
			user.Updated = userUpdated.Time
			user.IsFlagged = isFlagged.Bool
			user.SessionToken = sessionToken.String
			user.CSRFToken = csrfToken.String
			user.HashedPassword = hashedPassword.String

			models.UpdateTimeSince(&user)
			message.Sender = &user
		} else {
			message.Sender = nil
		}

		messages = append(messages, message)
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, fmt.Errorf("failed to commit transaction for GetChatMessages: %w", commitErr)
	}

	return messages, nil
}

// GetChatParticipantIDs returns all user IDs that are participants in the given chat
func (c *ChatModel) GetChatParticipantIDs(ctx context.Context, chatID models.UUIDField) ([]models.UUIDField, error) {
	query := `SELECT UserID FROM ChatUsers WHERE ChatID = ?`
	rows, err := c.DB.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chat participants: %w", err)
	}
	defer rows.Close()

	var participantIDs []models.UUIDField
	for rows.Next() {
		var userID models.UUIDField
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan participant ID: %w", err)
		}
		participantIDs = append(participantIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chat participants: %w", err)
	}

	return participantIDs, nil
}

// IsUserInChat checks if a user is a participant in the given chat
func (c *ChatModel) IsUserInChat(ctx context.Context, chatID, userID models.UUIDField) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM ChatUsers WHERE ChatID = ? AND UserID = ?)`
	var exists bool
	err := c.DB.QueryRowContext(ctx, query, chatID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is in chat: %w", err)
	}
	return exists, nil
}

// GetBuddyChatID returns the chat ID if a buddy chat exists between two users
func (c *ChatModel) GetBuddyChatID(ctx context.Context, user1ID, user2ID models.UUIDField) (models.UUIDField, error) {
	query := `
		SELECT DISTINCT c.ID
		FROM Chats c
		INNER JOIN ChatUsers cu1 ON c.ID = cu1.ChatID
		INNER JOIN ChatUsers cu2 ON c.ID = cu2.ChatID
		WHERE c.Type = 'buddy'
		AND cu1.UserID = ?
		AND cu2.UserID = ?
		LIMIT 1
	`
	var chatID models.UUIDField
	err := c.DB.QueryRowContext(ctx, query, user1ID, user2ID).Scan(&chatID)
	if err != nil {
		return models.ZeroUUIDField(), fmt.Errorf("failed to find buddy chat: %w", err)
	}
	return chatID, nil
}
