package handlers

import (
	"encoding/json"
	"fmt"
	mw "github.com/gary-norman/forum/internal/http/middleware"
	"log"
	"net/http"
	"regexp"

	"github.com/gary-norman/forum/internal/app"
	"github.com/gary-norman/forum/internal/colors"
	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/service"
	"github.com/gary-norman/forum/internal/view"
)

var (
	Colors, _ = colors.UseFlavor("Mocha")
	ErrorMsgs = models.CreateErrorMessages()
)

type AuthHandler struct {
	App     *app.App
	Session *SessionHandler
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := r.FormValue("register_user")
	email := r.FormValue("register_email")
	validEmail, _ := regexp.MatchString(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`, email)
	password := r.FormValue("register_password")
	if len(username) < 5 || len(username) > 16 {
		w.WriteHeader(http.StatusNotAcceptable)
		err := json.NewEncoder(w).Encode(map[string]any{
			"code":    http.StatusNotAcceptable,
			"message": "username must be between 5 and 16 characters",
		})
		if err != nil {
			models.LogErrorWithContext(ctx, "Failed to encode register response (username validation)", err)
			return
		}
		return
	}
	if !IsValidPassword(password) {
		w.WriteHeader(http.StatusNotAcceptable)
		err := json.NewEncoder(w).Encode(map[string]any{
			"code": http.StatusNotAcceptable,
			"message": "password must contain at least one number and one uppercase and lowercase letter," +
				"and at least 8 or more characters",
		})
		if err != nil {
			models.LogErrorWithContext(ctx, "Failed to encode register response (password validation)", err)
			return
		}
		return
	}
	if !validEmail {
		w.WriteHeader(http.StatusNotAcceptable)
		err := json.NewEncoder(w).Encode(map[string]any{
			"code":    http.StatusNotAcceptable,
			"message": "please enter a valid email address",
		})
		if err != nil {
			models.LogErrorWithContext(ctx, "Failed to encode register response (email validation)", err)
			return
		}
		return
	}
	_, ok, emailErr := h.App.Users.QueryUserEmailExists(ctx, email)
	if ok {
		w.WriteHeader(http.StatusConflict)
		encErr := json.NewEncoder(w).Encode(map[string]any{
			"code":    http.StatusConflict,
			"message": "an account is already registered to that email address",
			"body":    emailErr,
		})
		if encErr != nil {
			models.LogErrorWithContext(ctx, "Failed to encode register response (email exists)", encErr)
			return
		}
		return
	}
	_, ok, usernameErr := h.App.Users.QueryUserNameExists(ctx, username)
	if ok {
		w.WriteHeader(http.StatusConflict)
		encErr := json.NewEncoder(w).Encode(map[string]any{
			"code":    http.StatusConflict,
			"message": "an account is already registered to that username",
			"body":    usernameErr,
		})
		if encErr != nil {
			models.LogErrorWithContext(ctx, "Failed to encode register response (username exists)", encErr)
			return
		}
		return
	}

	user, err := service.NewUser(username, email, password)
	if err != nil {
		models.LogErrorWithContext(ctx, "Failed to create user %s", err, username)
	}

	if err := h.App.Users.Insert(
		ctx,
		user.ID,
		user.Username,
		user.Email,
		user.Avatar,
		user.Banner,
		user.Description,
		user.Usertype,
		user.SessionToken,
		user.CSRFToken,
		user.HashedPassword,
	); err != nil {
		models.LogErrorWithContext(ctx, "Failed to insert user into database", err)
		w.WriteHeader(http.StatusInternalServerError)
		encErr := json.NewEncoder(w).Encode(map[string]any{
			"code":    http.StatusInternalServerError,
			"message": "registration failed!",
		})
		if encErr != nil {
			models.LogErrorWithContext(ctx, "Failed to encode register error response", encErr)
			return
		}
	}

	type FormFields struct {
		Fields map[string][]string `json:"formValues"`
	}
	formFields := make(map[string][]string)
	for field, value := range r.Form {
		fieldName := field
		formFields[fieldName] = append(formFields[fieldName], value...)
	}
	// Send success response
	w.WriteHeader(http.StatusOK)
	encErr := json.NewEncoder(w).Encode(map[string]any{
		"code":    http.StatusOK,
		"message": "registration successful!",
		"body":    FormFields{Fields: formFields},
	})
	if encErr != nil {
		models.LogErrorWithContext(ctx, "Failed to encode register success response", encErr)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) GetWebsocketOTP(w http.ResponseWriter, r *http.Request) {
	// Verify user is authenticated
	_, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate OTP
	otp := h.App.Websocket.OTPs.NewOTP()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"otp": otp.Key,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Parse JSON from the request body
	var credentials struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Ephemeral bool   `json:"ephemeral"`
	}
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		view.RenderErrorPage(w, models.NotProcess("login"), 500, models.ParseError("request body", "Login", err))
		return
	}

	login := credentials.Username
	password := credentials.Password
	ephemeral := credentials.Ephemeral
	fmt.Printf(Colors.Peach+"Attempting login for "+Colors.Text+"%v\n"+Colors.Reset, login)
	fmt.Println(ErrorMsgs.Divider)

	user, getUserErr := h.App.Users.GetUserFromLogin(ctx, login, "login")
	if getUserErr != nil {
		// Respond with an unsuccessful login message
		w.Header().Set("Content-Type", "application/json")
		models.LogWarnWithContext(ctx, "User not found: %s", getUserErr, login)
		w.WriteHeader(http.StatusOK)
		encErr := json.NewEncoder(w).Encode(map[string]any{
			"code":    http.StatusUnauthorized,
			"message": "user not found",
		})
		if encErr != nil {
			models.LogErrorWithContext(ctx, "Failed to encode login response (user not found)", encErr)
			return
		}
		return
	}

	if models.CheckPasswordHash(password, user.HashedPassword) {
		// Set Session Token and CSRF Token cookies
		createCookiErr, expires := h.App.Cookies.CreateCookies(ctx, w, user, ephemeral)
		if createCookiErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			encErr := json.NewEncoder(w).Encode(map[string]any{
				"code":    http.StatusInternalServerError,
				"message": "failed to create cookies",
				"body":    fmt.Errorf(ErrorMsgs.Cookies, "create", createCookiErr),
			})
			if encErr != nil {
				models.LogErrorWithContext(ctx, "Failed to encode login response (cookie creation)", encErr)
				return
			}
			return
		}

		//adding OTP to a logged-in user for websocket authentication
		otp := h.App.Websocket.OTPs.NewOTP()

		// Respond with a successful login message
		models.LogInfoWithContext(ctx, ErrorMsgs.LoginSuccess, user.Username, expires)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encErr := json.NewEncoder(w).Encode(map[string]any{
			"code":    http.StatusOK,
			"message": fmt.Sprintf("Welcome, %s! Login successful.", user.Username),
			"otp":     otp.Key,
		})
		if encErr != nil {
			models.LogErrorWithContext(ctx, "Failed to encode login success response", encErr)
			return
		}
	} else {
		// Respond with an unsuccessful login message
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		encErr := json.NewEncoder(w).Encode(map[string]any{
			"code":    http.StatusUnauthorized,
			"message": "incorrect password",
		})
		if encErr != nil {
			models.LogErrorWithContext(ctx, "Failed to encode login failure response", encErr)
			return
		}
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Retrieve the cookie
	cookie, cookiErr := r.Cookie("username")
	if cookiErr != nil {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}
	username := cookie.Value
	if username == "" {
		models.LogWarnWithContext(ctx, "Logout aborted: no user is logged in")
		return
	}
	fmt.Printf(Colors.Peach+"Attempting logout for "+Colors.Text+"%v\n"+Colors.Reset, username)
	fmt.Println(ErrorMsgs.Divider)
	var user *models.User
	user, getUserErr := h.App.Users.GetUserByUsername(ctx, username, "logout")
	if getUserErr != nil {
		models.LogErrorWithContext(ctx, "Failed to get user %s for logout", getUserErr, username)
	}

	// Delete the Session Token and CSRF Token cookies
	delCookiErr := h.App.Cookies.DeleteCookies(ctx, w, user)
	if delCookiErr != nil {
		models.LogErrorWithContext(ctx, "Failed to delete cookies during logout", delCookiErr)
	}
	// send user confirmation
	models.LogInfoWithContext(ctx, "User %s logged out successfully", user.Username)
	encErr := json.NewEncoder(w).Encode(map[string]any{
		"code":    http.StatusOK,
		"message": "Logged out successfully!",
	})
	if encErr != nil {
		models.LogErrorWithContext(ctx, "Failed to encode logout success response", encErr)
		return
	}
}

// SECTION ------- routing handlers ----------

func (h *AuthHandler) Protected(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	login := r.FormValue("username")
	var user *models.User
	user, getUserErr := h.App.Users.GetUserFromLogin(ctx, login, "protected")
	if getUserErr != nil {
		models.LogErrorWithContext(ctx, "Failed to get user %s for protected route", getUserErr, login)
	}
	if authErr := h.Session.IsAuthenticated(r, user.Username); authErr != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	_, err := fmt.Fprintf(w, "CSRF Valildation successful! Welcome, %s", user.Username)
	if err != nil {
		models.LogErrorWithContext(ctx, "Failed to write protected route response for user %s", err, user.Username)
		return
	}
}
