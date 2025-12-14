package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gary-norman/forum/internal/models"
)

type CookieModel struct {
	DB *sql.DB
}

func (m *CookieModel) CreateCookies(w http.ResponseWriter, user *models.User, ephemeral bool) (error, time.Time) {
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

	if err := m.UpdateCookies(user, sessionToken, csrfToken, expires); err != nil {
		log.Printf(ErrorMsgs.Cookies, fmt.Sprintf("update user: %v", user.ID), err)
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
		log.Printf(ErrorMsgs.Cookies, "query", err)
	}
	defer rows.Close()

	var expire time.Time
	for rows.Next() {
		if err := rows.Scan(&expire); err != nil {
			log.Printf(ErrorMsgs.Cookies, "scan", err)
		}
	}

	// Get the Session Token from the request cookie
	st, err := r.Cookie("session_token")
	if err != nil {
		log.Printf(ErrorMsgs.Cookies, "query", err)
		return false
	}
	csrf, _ := r.Cookie("csrf_token")

	// Get the CSRF Token from the headers
	csrfToken := r.Header.Get("x-csrf-token")

	stColor, csrfColor := Colors.Red, Colors.Red
	stMatchString, csrfMatchString := "Failed!", "Failed!"
	if st.Value == user.SessionToken && time.Now().Before(expire) {
		stColor = Colors.Green
		stMatchString = "Success!"
		success = true
	} else {
		err := m.DeleteCookies(w, user)
		if err != nil {
			log.Printf("error deleting cookies")
		}
		success = false
	}
	if csrf.Value == csrfToken && csrfToken == user.CSRFToken {
		csrfColor = Colors.Green
		csrfMatchString = "Success!"
	}
	log.Printf(ErrorMsgs.KeyValuePair, "Cookie SessionToken", st.Value)
	log.Printf(ErrorMsgs.KeyValuePair, "User SessionToken", user.SessionToken)
	log.Printf(Colors.Blue+"Session token verficiation: "+stColor+"%v\n"+Colors.Reset, stMatchString)
	fmt.Println(ErrorMsgs.Divider)
	log.Printf(ErrorMsgs.KeyValuePair, "Cookie csrfToken", csrf.Value)
	log.Printf(ErrorMsgs.KeyValuePair, "Header csrfToken", csrfToken)
	log.Printf(ErrorMsgs.KeyValuePair, "User csrfToken", user.CSRFToken)
	log.Printf(Colors.Blue+"CSRF token verficiation: "+csrfColor+"%v\n"+Colors.Reset, csrfMatchString)

	return success
}

func (m *CookieModel) UpdateCookies(user *models.User, sessionToken, csrfToken string, expires time.Time) error {
	if m == nil || m.DB == nil {
		fmt.Printf(ErrorMsgs.UserModel, "UpdateCookies", user.Username)
		return errors.New("UserModel or DB is nil")
	}
	var stmt string
	fmt.Printf(Colors.Blue+"Updating DB Cookies for: "+Colors.Text+"%v\n"+Colors.Reset, user.Username)
	stmt = "UPDATE Users SET SessionToken = ?, CsrfToken = ?, CookiesExpire = ? WHERE Username = ?"
	result, err := m.DB.Exec(stmt, sessionToken, csrfToken, expires, user.Username)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	dbUpdated := "Failed!"
	dbUpdatedColor := Colors.Red
	if rows > 0 {
		dbUpdated = "Success!"
		dbUpdatedColor = Colors.Green
	}
	fmt.Printf(Colors.Blue+"Database update: "+dbUpdatedColor+"%v"+Colors.Reset+"\n", dbUpdated)

	return err
}

func (m *CookieModel) DeleteCookies(w http.ResponseWriter, user *models.User) error {
	expires := time.Now().Add(time.Hour - 1000)
	stmt := "UPDATE Users SET SessionToken = '', CsrfToken = '' WHERE Username = ?"
	result, err := m.DB.Exec(stmt, user.Username)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	dbUpdated := "Failed!"
	dbUpdatedColor := Colors.Red
	if rows > 0 {
		dbUpdated = "Success!"
		dbUpdatedColor = Colors.Green
	}
	log.Printf(Colors.Blue+"Database update: "+dbUpdatedColor+"%v"+Colors.Reset+"\n", dbUpdated)
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
	return err
}
