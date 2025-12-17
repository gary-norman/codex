// Package view contains the HTML rendering logic for the forum application.
package view

import (
	"context"
	"html/template"
	"path/filepath"

	"github.com/gary-norman/forum/internal/app"
	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/sqlite"
)

type TempHelper struct {
	App *app.App
}

var Template *template.Template

// reactionStatusWrapper wraps GetReactionStatus for template use
// Templates don't have access to request context, so we use background context
func (t *TempHelper) reactionStatusWrapper(authorID models.UUIDField, reactedPostID, reactedCommentID int64) (sqlite.ReactionStatus, error) {
	return t.App.Reactions.GetReactionStatus(context.Background(), authorID, reactedPostID, reactedCommentID)
}

// Init Function to initialise the custom template functions
func (t *TempHelper) Init() {
	tmplFiles1, _ := filepath.Glob("assets/templates/*.html")
	tmplFiles2, _ := filepath.Glob("assets/templates/*.tmpl")
	allFiles := append(tmplFiles1, tmplFiles2...)
	Template = template.Must(template.New("").Funcs(template.FuncMap{
		"compareAsInts":  compareAsInts,
		"debugPanic":     debugPanic,
		"decrement":      decrement,
		"dict":           dict,
		"fprint":         fprint,
		"increment":      increment,
		"isValZero":      isValZero,
		"not":            not,
		"or":             or,
		"printType":      printType,
		"random":         RandomInt,
		"reactionStatus": t.reactionStatusWrapper,
		"same":           checkSameName,
		"startsWith":     startsWith,
	}).ParseFiles(allFiles...))
}
