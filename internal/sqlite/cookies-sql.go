package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gary-norman/forum/internal/models"
)

type CookieModel struct {
	DB *sql.DB
}

var (
	dbUpdated                      string = "✘ Failed!"
	dbUpdatedColor                        = Colors.Red
	stColor, csrfColor                    = Colors.Red, Colors.Red
	stMatchString, csrfMatchString        = "✘ Failed!", "✘ Failed!"
	successFail                           = fmt.Sprintf(" --> %s%s%s", dbUpdatedColor, dbUpdated, Colors.Reset)
)

func (m *CookieModel) CreateCookies(ctx context.Context, w http.ResponseWriter, user *models.User, ephemeral bool) (error, time.Time) {
	sessionToken := models.GenerateToken(32)
	csrfToken := models.GenerateToken(32)
	var expires time.Time
	if ephemeral {
		expires = time.Now().Add(24 * time.Hour)
	} else {
		expires = time.Now().AddDate(0, 3, 0)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expires,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "username",
		Value:    user.Username,
		Expires:  expires,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Expires:  expires,
		HttpOnly: false,
	})

	if err := m.UpdateCookies(ctx, user, sessionToken, csrfToken, expires); err != nil {
		models.LogErrorWithContext(ctx, "Failed to update cookies for user", err, "UserID:", user.ID)
		return err, time.Now()
	}
	return nil, expires
}

func (m *CookieModel) QueryCookies(w http.ResponseWriter, r *http.Request, user *models.User) bool {
	var success bool
	ctx := r.Context()
	stmt := "SELECT CookiesExpire FROM Users WHERE Username = ?"
	rows, err := m.DB.QueryContext(ctx, stmt, user.Username)
	if err != nil {
		models.LogErrorWithContext(ctx, "Failed to query cookie expiration", err, "Username:", user.Username)
		return false
	}
	defer rows.Close()

	var expire time.Time
	for rows.Next() {
		if err := rows.Scan(&expire); err != nil {
			models.LogErrorWithContext(ctx, "Failed to scan cookie expiration row", err)
		}
	}

	// Get the Session Token from the request cookie
	st, err := r.Cookie("session_token")
	if err != nil {
		models.LogErrorWithContext(ctx, "Failed to get session_token cookie", err)
		return false
	}
	csrf, _ := r.Cookie("csrf_token")

	// Get the CSRF Token from the headers
	csrfToken := r.Header.Get("x-csrf-token")

	if st.Value == user.SessionToken && time.Now().Before(expire) {
		stColor = Colors.Green
		stMatchString = "Success!"
		success = true
	} else {
		err := m.DeleteCookies(ctx, w, user)
		if err != nil {
			models.LogErrorWithContext(ctx, "Failed to delete expired cookies", err, "Username:", user.Username)
		}
		success = false
	}
	if csrf.Value == csrfToken && csrfToken == user.CSRFToken {
		csrfColor = Colors.Green
		csrfMatchString = "Success!"
	}
	models.LogInfoWithContext(ctx, "Cookie SessionToken: %s", st.Value)
	models.LogInfoWithContext(ctx, "User SessionToken: %s", user.SessionToken)
	models.LogInfoWithContext(ctx, "Session token verification: %s%s%s", stColor, stMatchString, Colors.Reset)
	models.LogInfoWithContext(ctx, "Cookie CSRF token: %s", csrf.Value)
	models.LogInfoWithContext(ctx, "Header CSRF token: %s", csrfToken)
	models.LogInfoWithContext(ctx, "User CSRF token: %s", user.CSRFToken)
	models.LogInfoWithContext(ctx, "CSRF token verification: %s%s%s", csrfColor, csrfMatchString, Colors.Reset)

	return success
}

func (m *CookieModel) UpdateCookies(ctx context.Context, user *models.User, sessionToken, csrfToken string, expires time.Time) error {
	if m == nil || m.DB == nil {
		models.LogErrorWithContext(ctx, "CookieModel or DB is nil in UpdateCookies", nil, "Username:", user.Username)
		return errors.New("UserModel or DB is nil in UpdateCookies")
	}
	var stmt string
	fmt.Printf(Colors.Blue+"Updating DB Cookies for: "+Colors.Text+"%v\n"+Colors.Reset, user.Username)
	stmt = "UPDATE Users SET SessionToken = ?, CsrfToken = ?, CookiesExpire = ? WHERE Username = ?"
	result, err := m.DB.ExecContext(ctx, stmt, sessionToken, csrfToken, expires, user.Username)
	if err != nil {
		return fmt.Errorf("failed to update cookies for user %s: %w", user.Username, err)
	}
	rows, _ := result.RowsAffected()
	if rows > 0 {
		dbUpdated = "✔ Success!"
		dbUpdatedColor = Colors.Green
	}
	models.LogInfoWithContext(ctx, "Updating cookies for user: %s%s", user.Username, successFail)

	return nil
}

func (m *CookieModel) DeleteCookies(ctx context.Context, w http.ResponseWriter, user *models.User) error {
	expires := time.Now().Add(time.Hour - 1000)
	stmt := "UPDATE Users SET SessionToken = '', CsrfToken = '' WHERE Username = ?"
	result, err := m.DB.ExecContext(ctx, stmt, user.Username)
	if err != nil {
		return fmt.Errorf("failed to delete cookies for user %s: %w", user.Username, err)
	}
	rows, _ := result.RowsAffected()
	if rows > 0 {
		dbUpdated = "✔ Success!"
		dbUpdatedColor = Colors.Green
	}
	models.LogInfoWithContext(ctx, "Deleting cookies for user: %s%s", user.Username, successFail)
	// Set Session, Username, and CSRF Token cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  expires,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "username",
		Value:    "",
		Expires:  expires,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Expires:  expires,
		HttpOnly: false,
	})
	return nil
}
