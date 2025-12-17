package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gary-norman/forum/internal/models"
)

type UserModel struct {
	DB *sql.DB
}

func CountUsers(ctx context.Context, db *sql.DB) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ID`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Insert adds a new user to the database
func (m *UserModel) Insert(ctx context.Context, id models.UUIDField, username, email, avatar, banner, description, userType, sessionToken, crsfToken, password string) error {
	// Note: Direct Exec() is more efficient than Prepare() for single-use queries
	query := "INSERT INTO Users (ID, Username, EmailAddress, Avatar, Banner, Description, UserType, Created, IsFlagged, SessionToken, CsrfToken, HashedPassword) VALUES (?, ?, ?, ?, ?, ?, ?, DateTime('now'), 0, ?, ?, ?)"

	_, err := m.DB.ExecContext(ctx, query, id, username, email, avatar, banner, description, userType, sessionToken, crsfToken, password)
	if err != nil {
		return fmt.Errorf("failed to insert user %s: %w", username, err)
	}

	models.LogInfo("User created: %s", username)
	return nil
}

func (m *UserModel) Edit(ctx context.Context, user *models.User) error {
	query := "UPDATE Users SET Username = ?, EmailAddress = ?, HashedPassword = ?, SessionToken = ?, CsrfToken = ?, Avatar = ?, Banner = ?, Description = ? WHERE ID = ?"

	result, err := m.DB.ExecContext(ctx, query, user.Username, user.Email, user.HashedPassword, user.SessionToken, user.CSRFToken, user.Avatar, user.Banner, user.Description, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user %s: %w", user.Username, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		models.LogWarn("User update affected 0 rows: %s", user.Username)
	}

	models.LogInfo("User updated: %s", user.Username)
	return nil
}

func (m *UserModel) Delete(ctx context.Context, user *models.User) error {
	query := "DELETE FROM Users WHERE ID = ?"

	result, err := m.DB.ExecContext(ctx, query, user.ID)
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", user.Username, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		models.LogWarn("User delete affected 0 rows: %s", user.Username)
	}

	models.LogInfo("User deleted: %s", user.Username)
	return nil
}

func (m *UserModel) GetUserFromLogin(ctx context.Context, login, calledBy string) (*models.User, error) {
	if m == nil || m.DB == nil {
		return nil, fmt.Errorf("error connecting to database called by: %s", calledBy)
	}
	username, email := login, login
	var loginType string
	usernameQuery, ok, _ := m.QueryUserNameExists(ctx, username)
	if ok {
		loginType = usernameQuery
	}
	emailQuery, ok, _ := m.QueryUserEmailExists(ctx, email)
	if ok {
		loginType = emailQuery
	}
	switch loginType {
	case "username":
		user, err := m.GetUserByUsername(ctx, username, "GetUserFromLogin")
		if err != nil {
			return nil, fmt.Errorf("failed to get user by username: %w", err)
		} else {
			models.LogInfo("Successfully found user by username: %s", user.Username)
			return user, nil
		}
	case "email":
		user, err := m.GetUserByEmail(ctx, email, "GetUserFromLogin")
		if err != nil {
			return nil, fmt.Errorf("failed to get user by email: %w", err)
		} else {
			models.LogInfo("Successfully found user by email: %s", user.Username)
			return user, nil
		}
	default:
		return nil, fmt.Errorf("user: %v not found", login)
	}
}

func (m *UserModel) QueryUserNameExists(ctx context.Context, username string) (string, bool, error) {
	if m == nil || m.DB == nil {
		err := fmt.Errorf("error connecting to database: %s", "QueryUserNameExists")
		return "", false, err

	}
	var count int
	queryErr := m.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM Users WHERE Username = ?", username).Scan(&count)
	if queryErr != nil {
		return "", false, fmt.Errorf("failed to query user by username: %w", queryErr)
	}
	if count > 0 {
		return "username", true, nil
	}
	return "", false, nil
}

func (m *UserModel) QueryUserEmailExists(ctx context.Context, email string) (string, bool, error) {
	if m == nil || m.DB == nil {
		err := fmt.Errorf("error connecting to database: %s", "QueryUserEmailExists")
		return "", false, err
	}
	var count int
	queryErr := m.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM Users WHERE EmailAddress = ?", email).Scan(&count)
	if queryErr != nil {
		return "", false, fmt.Errorf("failed to query user by email: %w", queryErr)
	}
	if count > 0 {
		return "email", true, nil
	}
	return "", false, nil
}

// TODO unify these functions to accept parameters

func (m *UserModel) GetUserByUsername(ctx context.Context, username, calledBy string) (*models.User, error) {
	username = strings.TrimSpace(username)
	if m == nil || m.DB == nil {
		return nil, fmt.Errorf("database not initialized in GetUserByUsername for %s", username)
	}

	query := "SELECT ID, Username, EmailAddress, Avatar, Banner, Description, Usertype, Created, Updated, IsFlagged, SessionToken, CSRFToken, HashedPassword FROM Users WHERE Username = ? LIMIT 1"
	var user models.User

	err := m.DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Avatar,
		&user.Banner,
		&user.Description,
		&user.Usertype,
		&user.Created,
		&user.Updated,
		&user.IsFlagged,
		&user.SessionToken,
		&user.CSRFToken,
		&user.HashedPassword)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s: %w", username, err)
		}
		return nil, fmt.Errorf("failed to get user by username %s: %w", username, err)
	}

	return &user, nil
}

func (m *UserModel) GetUserByEmail(ctx context.Context, email, calledBy string) (*models.User, error) {
	email = strings.TrimSpace(email)
	if m == nil || m.DB == nil {
		return nil, fmt.Errorf("database not initialized in GetUserByEmail for %s", email)
	}

	query := "SELECT ID, HashedPassword, EmailAddress FROM Users WHERE EmailAddress = ? LIMIT 1"
	var user models.User

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.HashedPassword,
		&user.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found by email: %s: %w", email, err)
		}
		return nil, fmt.Errorf("failed to get user by email %s: %w", email, err)
	}

	return &user, nil
}

func (m *UserModel) GetUserByID(ctx context.Context, ID models.UUIDField) (models.User, error) {
	stmt := "SELECT ID, Username, EmailAddress, Avatar, Banner, Description, Usertype, Created, Updated, IsFlagged, SessionToken, CSRFToken, HashedPassword FROM Users WHERE ID = ?"
	row := m.DB.QueryRowContext(ctx, stmt, ID)
	u := models.User{}
	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Avatar,
		&u.Banner,
		&u.Description,
		&u.Usertype,
		&u.Created,
		&u.Updated,
		&u.IsFlagged,
		&u.SessionToken,
		&u.CSRFToken,
		&u.HashedPassword)
	if err != nil {
		return u, fmt.Errorf("failed to get user by ID %s: %w", ID, err)
	}
	models.UpdateTimeSince(&u)
	return u, nil
}

// TODO accept an interface for any given value
func isValidUserColumn(column string) bool {
	validColumns := map[string]bool{
		"ID":             true,
		"Username":       true,
		"EmailAddress":   true,
		"HashedPassword": true,
		"SessionToken":   true,
		"CsrfToken":      true,
		"Avatar":         true,
		"Banner":         true,
		"Description":    true,
		"UserType":       true,
		"Created":        true,
		"Updated":        true,
		"IsFlagged":      true,
	}
	return validColumns[column]
}

// GetSingleUserValue returns the string of the column specified in output, which should be entered in all lower case
func (m *UserModel) GetSingleUserValue(ctx context.Context, ID models.UUIDField, searchColumn, outputColumn string) (string, error) {
	if !isValidUserColumn(searchColumn) {
		return "", fmt.Errorf("invalid searchColumn name: %s", searchColumn)
	}
	stmt := fmt.Sprintf(
		"SELECT ID, Username, EmailAddress, Avatar, Banner, Description, Usertype, Created, IsFlagged, SessionToken, CSRFToken, HashedPassword FROM Users WHERE %s = ?",
		searchColumn,
	)
	rows, queryErr := m.DB.QueryContext(ctx, stmt, ID)
	if queryErr != nil {
		return "", fmt.Errorf("failed to query user for column %s: %w", searchColumn, queryErr)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()
	var user models.User
	if rows.Next() {
		if scanErr := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Avatar, &user.Banner, &user.Description, &user.Usertype,
			&user.Created, &user.IsFlagged, &user.SessionToken, &user.CSRFToken, &user.HashedPassword); scanErr != nil {
			return "", scanErr
		}
	} else {
		return "", fmt.Errorf("no user found")
	}

	// Map searchColumn names to their values
	fields := map[string]any{
		"id":             user.ID,
		"username":       user.Username,
		"email":          user.Email,
		"hashedPassword": user.HashedPassword,
		"sessionToken":   user.SessionToken,
		"csrfToken":      user.CSRFToken,
		"avatar":         user.Avatar,
		"banner":         user.Banner,
		"description":    user.Description,
		"usertype":       user.Usertype,
		"created":        user.Created,
		"updated":        user.Updated,
		"isFlagged":      user.IsFlagged,
	}

	// Check if outputColumn exists in the map
	value, exists := fields[outputColumn]
	if !exists {
		return "", fmt.Errorf("invalid search Column name: %s", outputColumn)
	}

	// Convert the value to a string (handling different types)
	outputValue := fmt.Sprintf("%v", value)
	return outputValue, nil
}

func (m *UserModel) All(ctx context.Context) ([]*models.User, error) {
	stmt := "SELECT ID, Username, EmailAddress, Avatar, Banner, Description, Usertype, Created, Updated, IsFlagged, SessionToken, CSRFToken, HashedPassword FROM Users ORDER BY ID DESC"
	rows, queryErr := m.DB.QueryContext(ctx, stmt)
	if queryErr != nil {
		return nil, fmt.Errorf("failed to query all users: %w", queryErr)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()

	users := make([]*models.User, 0)
	for rows.Next() {
		p, err := parseUserRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error parsing row: %w", err)
		}
		users = append(users, p)
	}
	return users, nil
}

func parseUserRows(rows *sql.Rows) (*models.User, error) {
	var user models.User

	if err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Avatar,
		&user.Banner,
		&user.Description,
		&user.Usertype,
		&user.Created,
		&user.Updated,
		&user.IsFlagged,
		&user.SessionToken,
		&user.CSRFToken,
		&user.HashedPassword,
	); err != nil {
		return nil, fmt.Errorf("Error parsing UserRows: %w", err)
	}
	models.UpdateTimeSince(&user)
	return &user, nil
}

func parseUserRow(row *sql.Row) (*models.User, error) {
	var user models.User

	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Avatar,
		&user.Banner,
		&user.Description,
		&user.Usertype,
		&user.Created,
		&user.Updated,
		&user.IsFlagged,
		&user.SessionToken,
		&user.CSRFToken,
		&user.HashedPassword,
	); err != nil {
		return nil, fmt.Errorf("Error parsing UserRow: %w", err)
	}
	models.UpdateTimeSince(&user)
	return &user, nil
}
