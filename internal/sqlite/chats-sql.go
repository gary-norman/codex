package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type ChatModel struct {
	DB *sql.DB
}

func (c *ChatModel) CreateChat(chatType, name string, groupID, buddyID models.UUIDField) (models.UUIDField, error) {
	chatID := models.NewUUIDField()
	query := "INSERT INTO Chats (ID, Type, Name, GroupID, BuddyID, Created) VALUES (?, ?, ?, ?, ?, DateTime('now'))"
	_, err := c.DB.Exec(query, chatID, chatType, name, groupID, buddyID)
	if err != nil {
		return models.UUIDField{}, fmt.Errorf("failed to insert chat: %w", err)
	}

	return chatID, nil
}

func (c *ChatModel) CreateChatMessage(chatID, userID models.UUIDField, message string) (models.UUIDField, error) {
	messageID := models.NewUUIDField()
	query := "INSERT INTO Messages (ID, ChatID, UserID, Created, Content) VALUES (?, ?, ?, DateTime('now'), ?)"
	_, err := c.DB.Exec(query, messageID, chatID, userID, message)
	if err != nil {
		return models.UUIDField{}, fmt.Errorf("failed to insert message: %w", err)
	}

	return messageID, nil
}

func (c *ChatModel) AttachUserToChat(chatID, userID models.UUIDField) error {
	query := "INSERT INTO ChatUsers (ChatID, UserID) VALUES (?, ?)"
	_, err := c.DB.Exec(query, chatID, userID)
	if err != nil {
		return fmt.Errorf("failed to attach user to chat: %w", err)
	}

	return nil
}

func (c *ChatModel) GetUserChatIDs(userID models.UUIDField) ([]models.UUIDField, error) {
	query := `SELECT ChatID FROM ChatUsers WHERE UserID = ?`
	rows, err := c.DB.Query(query, userID)
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
func (c *ChatModel) GetChat(chatID models.UUIDField) (*models.Chat, error) {
	query := "SELECT ID, Type, Name, Created, LastActive, GroupID, BuddyID FROM Chats WHERE ID = ?"
	row := c.DB.QueryRow(query, chatID)

	var chat models.Chat
	var buddyID, groupID models.NullableUUIDField

	err := row.Scan(&chat.ID, &chat.ChatType, &chat.Name, &chat.Created, &chat.LastActive, &groupID, &buddyID)
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

	return &chat, nil
}

// GetUserChats retrieves all chats for a specific user
func (c *ChatModel) GetUserChats(userID models.UUIDField) ([]models.Chat, error) {
	query := `
		SELECT c.ID, c.Type, c.Name, c.Created, c.LastActive, c.GroupID, c.BuddyID
		FROM Chats c
		INNER JOIN ChatUsers cu ON c.ID = cu.ChatID
		WHERE cu.UserID = ?
		ORDER BY c.LastActive DESC
	`

	rows, err := c.DB.Query(query, userID)
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
func (c *ChatModel) GetChatMessages(chatID models.UUIDField) ([]models.ChatMessage, error) {
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

	rows, err := c.DB.Query(query, chatID)
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

	return messages, nil
}
