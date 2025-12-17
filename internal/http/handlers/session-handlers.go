package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gary-norman/forum/internal/app"
	"github.com/gary-norman/forum/internal/models"
)

type SessionHandler struct {
	App *app.App
}

var (
	authUser                       string = "✘ Failed!"
	authUserColor                         = Colors.Red
	stColor, csrfColor                    = Colors.Red, Colors.Red
	stMatchString, csrfMatchString        = "✘ Failed!", "✘ Failed!"
	successFail                           = fmt.Sprintf("Authorise user %s%s%s for user: ", authUserColor, authUser, Colors.Reset)
)

func (s *SessionHandler) IsAuthenticated(r *http.Request, username string) error {
	ctx := r.Context()
	var user *models.User
	user, getUserErr := s.App.Users.GetUserByUsername(ctx, username, "isAuthenticated")
	if getUserErr != nil {
		return fmt.Errorf(ErrorMsgs.NotFound, username, "isAuthenticated", getUserErr)
	}
	// Get the Session Token from the request cookie
	st, err := r.Cookie("session_token")
	if st == nil {
		return errors.New("no session token")
	}
	if err != nil || st.Value == "" || st.Value != user.SessionToken {
		// fmt.Printf(ErrorMsgs.KeyValuePair, "Cookie SessionToken", st.Value)
		// fmt.Printf(ErrorMsgs.KeyValuePair, "Error", err)
		// fmt.Printf(ErrorMsgs.KeyValuePair, "User SessionToken", user.SessionToken)
		return fmt.Errorf("authentication failed: %w", err)
	}
	// csrf, _ := r.Cookie("csrf_token")

	// Get the CSRF Token from the headers
	csrfToken := r.Header.Get("x-csrf-token")
	// fmt.Printf(ErrorMsgs.KeyValuePair, "Header", r.Header)
	if csrfToken == "" || csrfToken != user.CSRFToken {
		authErr := fmt.Errorf("%s%s", successFail, user.Username)
		models.LogErrorWithContext(ctx, "CSRF token mismatch for user: %s", authErr, user.Username)
		return authErr
	}
	authUser = "✔ Success!"
	authUserColor = Colors.Green
	models.LogInfoWithContext(ctx, "CSRF token match for user: %s", successFail, user.Username)
	return nil
}
