package models

import (
	"fmt"
	"reflect"

	"github.com/gary-norman/forum/internal/colors"
)

type Errors struct {
	// Close: Unable to close %v called by %v with error: %v
	Close string
	// ConnClose: Unable to close connection to %v called by %v
	ConnClose string
	// ConnInit: Unable to initialise connection %v called by %v
	ConnInit string
	// ConnConn: Unable to connect to %v called by %v(%v)
	ConnConn string
	// ConnSuccess: Server listening on %v - success!
	ConnSuccess string
	// DBSuccess: Database connected: %v v%v - success!
	DBSuccess string
	// Convert: Unable to convert %v called by %v with error: %v
	Convert string
	// Cookies: Unable to %v cookies with error: %v
	Cookies string
	// CreateFile: Unable to create file %v called by %v with error: %v
	CreateFile string
	// Delete: Unable to delete %v called by %v with error: %v
	Delete string
	// Divider: Visual divider line (no placeholders)
	Divider string
	// Edit: Unable to edit %v called by %v with error: %v
	Edit string
	// Encode: Unable to encode JSON called by %v with error: %v
	Encode string
	// Execute: Unable to execute template with error: %v
	Execute string
	// Fetch: Unable to fetch %v called by %v with error: %v
	Fetch string
	// Generic: %s: %v called by %v with error: %v
	Generic string
	// Insert: Unable to insert %v called by %v with error: %v
	Insert string
	// KeyValuePair: %v: %v (blue/white pair)
	KeyValuePair string
	// LoginSuccess: Succesful login for user %s, expires: %v
	LoginSuccess string
	// Login: Unable to login %v called by %v with error: %v
	Login string
	// NoRows: No rows returned for %v called by %v with error: %v
	NoRows string
	// NotFound: Unable to find %v called by %v with error: %v
	NotFound string
	// Open: Unable to open %v called by %v with error: %v
	Open string
	// Parse: Unable to parse %v called by %v with error: %v
	Parse string
	// Post: Unable to post with error: %v
	Post string
	// Comment: Unable to Comment with error: %v
	Comment string
	// Printf: Unable to print with error: %v
	Printf string
	// Protected: CSRF validation failed for user %v with error: %v
	Protected string
	// Query: Unable to query %v called by %v with error: %v
	Query string
	// Read: Unable to read %v called by %v with error: %v
	Read string
	// RetrieveFile: Unable to retrieve file %v called by %v with error: %v
	RetrieveFile string
	// Register: Unable to register with error: %v
	Register string
	// SaveFile: Unable to save file %v to %v called by %v with error: %v
	SaveFile string
	// Shutdown: HTTP shutdown error: %v
	Shutdown string
	// Unmarshal: Unable to unmarshal %v with error: %v
	Unmarshal string
	// Update: Unable to update %v called by %v with error: %v
	Update string
	// UserModel: Usermodel or DB called in %v for %v is nil
	UserModel string
	// Write: Unable to write to %v called by %v
	Write string
}

type Message struct {
	Name, Text string
}

func JSONError(messageStruct TemplateData) {
	ErrorMsgs := CreateErrorMessages()
	val := reflect.ValueOf(messageStruct)
	typ := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)
		if fieldType.Name == "CurrentUser" {
			continue
		}
		if fieldType.Name == "Posts" {
			// fmt.Printf(ErrorMsgs.KeyValuePair, "Number of posts", len(field.Interface().([]PostWithWrapping)))
			continue
		}
		if fieldType.Name == "Images" {
			// fmt.Printf(ErrorMsgs.KeyValuePair, "Number of images", len(field.Interface().([]Image)))
			continue
		}
		if fieldType.Name == "Comments" {
			// fmt.Printf(ErrorMsgs.KeyValuePair, "Number of comments", len(field.Interface().([]Comment)))
			continue
		}
		if fieldType.Name == "Reactions" {
			// fmt.Printf(ErrorMsgs.KeyValuePair, "Number of reactions", len(field.Interface().([]Reaction)))
			continue
		}
		fmt.Printf(ErrorMsgs.KeyValuePair, fieldType.Name, field.Interface())
	}
}

func JSONPost(messageStruct any) {
	ErrorMsgs := CreateErrorMessages()
	val := reflect.ValueOf(messageStruct)
	typ := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)
		fmt.Printf(ErrorMsgs.KeyValuePair, fieldType.Name, field.Interface())
	}
}

var Colors, colorsErr = colors.UseFlavor("Mocha")

// Helper to generate standard error messages
func errMsg(action, subject, caller, err string) string {
	msg := Colors.Red + action + " " + Colors.Text + subject
	if caller != "" {
		msg += Colors.Blue + " called by " + Colors.Text + caller
	}
	if err != "" {
		msg += Colors.Blue + " with error: " + Colors.Red + err
	}
	return msg + Colors.Reset
}

// KeyValue helper
func kvMsg(key, value string) string {
	return Colors.Blue + key + ": " + Colors.Text + value + "\n"
}

// CreateErrorMessages initializes and returns an Errors struct with formatted messages
// set any field to "" to avoid it
func CreateErrorMessages() *Errors {
	if colorsErr != nil {
		fmt.Println(fmt.Errorf("error loading colors: %v", colorsErr))
	}
	return &Errors{
		Close:        errMsg("Unable to close", "%v", "%v", "%v"),
		ConnConn:     errMsg("Unable to connect to", "%v", "%v(%v)", ""),
		ConnClose:    errMsg("Unable to close connection to", "%v", "%v", ""),
		ConnInit:     errMsg("Unable to initialise connection", "%v", "%v", ""),
		ConnSuccess:  Colors.Blue + "Server listening on " + Colors.Text + "%v " + Colors.CodexGreen + "- success!\n" + Colors.Reset,
		DBSuccess:    Colors.Blue + "Database connected: " + Colors.Text + "%v " + Colors.CodexPink + "%v " + Colors.Green + "- success!\n" + Colors.Reset,
		Convert:      errMsg("Unable to convert", "%v", "%v", "%v"),
		Cookies:      errMsg("Unable to", "%v cookies", "", "%v"),
		CreateFile:   errMsg("Unable to create file", "%v", "%v", "%v"),
		Delete:       errMsg("Unable to delete", "%v", "%v", "%v"),
		Divider:      Colors.CodexPink + "-------------------------------------------------------" + Colors.Reset,
		Edit:         errMsg("Unable to edit", "%v", "%v", "%v"),
		Encode:       errMsg("Unable to encode JSON", "", "%v", "%v"),
		Execute:      errMsg("Unable to execute template", "", "", "%v"),
		Fetch:        errMsg("Unable to fetch", "%v", "%v", "%v"),
		Generic:      errMsg("%s", "%v", "%v", "%v"),
		Insert:       errMsg("Unable to insert", "%v", "%v", "%v"),
		KeyValuePair: kvMsg("%v", "%v"),
		Login:        errMsg("Unable to login", "%v", "%v", "%v"),
		LoginSuccess: errMsg("Successful login for user", "%s", "expires:", "%v"),
		NoRows:       errMsg("No rows returned for", "%v", "%v", "v"),
		NotFound:     errMsg("Unable to find", "%v", "%v", "%v"),
		Open:         errMsg("Unable to open", "%v", "%v", "%v"),
		Parse:        errMsg("Unable to parse", "%v", "%v", "%v"),
		Post:         errMsg("Unable to post", "", "", "%v"),
		Comment:      errMsg("Unable to comment", "", "", "%v"),
		Printf:       errMsg("Unable to print", "", "", "%v"),
		Protected:    errMsg("CSRF validation failed for user", "%v", "", "%v"),
		Query:        errMsg("Unable to query", "%v", "%v", "%v"),
		Read:         errMsg("Unable to read", "%v", "%v", "%v"),
		RetrieveFile: errMsg("Unable to retrieve file", "%v", "%v", "%v"),
		Register:     errMsg("Unable to register", "", "", "%v"),
		SaveFile:     errMsg("Unable to save file", "%v to %v", "%v", "%v"),
		Shutdown:     errMsg("HTTP shutdown error", "%v", "", ""),
		Update:       errMsg("Unable to update", "%v", "%v", "%v"),
		UserModel:    errMsg("Usermodel or DB called in", "%v for %v", "", ""),
		Write:        errMsg("Unable to write to", "%v", "%v", ""),
	}
}
