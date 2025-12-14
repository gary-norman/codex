package handlers

import (
	"fmt"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gary-norman/forum/internal/models"
)

func IsValidPassword(password string) bool {
	// At least 8 characters
	if len(password) < 8 {
		return false
	}
	// At least one digit
	hasDigit, _ := regexp.MatchString(`[0-9]`, password)
	if !hasDigit {
		return false
	}
	// At least one lowercase letter
	hasLower, _ := regexp.MatchString(`[a-z]`, password)
	if !hasLower {
		return false
	}
	// At least one uppercase letter
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	return hasUpper
}

func GetTimeSince(created time.Time) string {
	now := time.Now()
	hours := now.Sub(created).Hours()
	var timeSince string
	if hours > 24 {
		timeSince = fmt.Sprintf("%.0f days ago", hours/24)
	} else if hours > 1 {
		timeSince = fmt.Sprintf("%.0f hours ago", hours)
	} else if minutes := now.Sub(created).Minutes(); minutes > 1 {
		timeSince = fmt.Sprintf("%.0f minutes ago", minutes)
	} else {
		timeSince = "just now"
	}
	return timeSince
}

func GetRandomChannel(channelSlice []models.Channel) models.Channel {
	if len(channelSlice) == 0 {
		return models.Channel{}
	}
	rndInt := rand.IntN(len(channelSlice))
	channel := channelSlice[rndInt]
	return channel
}

func GetRandomUser(userSlice []*models.User) *models.User {
	if len(userSlice) == 0 {
		return &models.User{}
	}
	rndInt := rand.IntN(len(userSlice))
	user := userSlice[rndInt]
	return user
}

func GetFileName(r *http.Request, fileFieldName, calledBy, imageType string) string {
	// Limit the size of the incoming file to prevent memory issues
	parseErr := r.ParseMultipartForm(10 << 20) // Limit upload size to 10MB
	if parseErr != nil {
		models.LogError("Failed to parse multipart form in %s", parseErr, calledBy)
		return "noimage"
	}
	// Retrieve the file from form data
	file, handler, retrieveErr := r.FormFile(fileFieldName)
	if retrieveErr != nil {
		models.LogError("Failed to retrieve file in %s", retrieveErr, calledBy)
		return "noimage"
	}
	defer func(file multipart.File) {
		closeErr := file.Close()
		if closeErr != nil {
			models.LogError("Failed to close file in %s", closeErr, calledBy)
		}
	}(file)
	// Create a file in the server's local storage
	renamedFile := renameFileWithUUID(handler.Filename)
	models.LogInfo("Saving file: %s", renamedFile)
	dst, createErr := os.Create("db/userdata/images/" + imageType + "-images/" + renamedFile)
	if createErr != nil {
		models.LogError("Failed to create file in %s", createErr, calledBy)
		return ""
	}
	defer func(dst *os.File) {
		closeErr := dst.Close()
		if closeErr != nil {
			models.LogError("Failed to close destination file in %s", closeErr, calledBy)
		}
	}(dst)
	// Copy the uploaded file data to the server's file
	_, copyErr := io.Copy(dst, file)
	if copyErr != nil {
		models.LogError("Failed to save file in %s", copyErr, calledBy)
		return ""
	}
	return renamedFile
}

func renameFileWithUUID(oldFilePath string) string {
	// Generate a new UUID
	newUUID := models.GenerateToken(16)

	// Split the file name into its base and extension
	base := filepath.Base(oldFilePath)
	ext := filepath.Ext(base)
	// base = base[:len(base)-len(ext)]

	// Create the new file name with the UUID and extension
	newFilePath := filepath.Join(filepath.Dir(oldFilePath), newUUID+ext)

	return newFilePath
}
