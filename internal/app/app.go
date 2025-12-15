// Package app sets up the application.
package app

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/gary-norman/forum/internal/http/websocket"
	"log"
	"os"
	"strings"

	"github.com/gary-norman/forum/internal/colors"
	"github.com/gary-norman/forum/internal/db"
	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/sqlite"
)

type Config struct {
	DBType     string
	DBDriver   string
	DBEnv      string
	DBPath     string
	SchemaPath string
	ImagePath  string
}

var (
	Colors, _ = colors.UseFlavor("Mocha")
	ErrorMsgs = models.CreateErrorMessages()
)

// LoadEnv reads .env and sets os.Environ()
func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines or comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		os.Setenv(key, val)
	}

	return scanner.Err()
}

// InitConfig loads .env and builds the Config
func initConfig() *Config {
	if err := loadEnv("./.env"); err != nil {
		log.Fatalf("❌ failed to load .env: %v", err)
	}

	cfg := &Config{
		DBType:     "SQLite",
		DBDriver:   "sqlite3",
		DBEnv:      os.Getenv("DB_ENV"),
		DBPath:     os.Getenv("DB_PATH"),
		SchemaPath: "./migrations/001_schema.sql",
		ImagePath:  "/db/userdata/images/",
	}

	if cfg.DBEnv == "" || cfg.DBPath == "" {
		log.Fatal(Colors.Red + "❌ DB_ENV or DB_PATH missing" + Colors.Reset + "— run" + Colors.CodexPink + "`make configure`" + Colors.Reset + "first")
	}

	fmt.Printf(Colors.CodexGreen+"✓ "+Colors.Reset+"using "+Colors.CodexPink+"%s"+Colors.Reset+" database: "+Colors.CodexGreen+"%s\n", cfg.DBEnv, cfg.DBPath)
	return cfg
}

type App struct {
	DB          *sql.DB // Store DB reference for cleanup
	Users       *sqlite.UserModel
	Posts       *sqlite.PostModel
	Reactions   *sqlite.ReactionModel
	Saved       *sqlite.SavedModel
	Mods        *sqlite.ModModel
	Comments    *sqlite.CommentModel
	Images      *sqlite.ImageModel
	Channels    *sqlite.ChannelModel
	Flags       *sqlite.FlagModel
	Loyalty     *sqlite.LoyaltyModel
	Memberships *sqlite.MembershipModel
	Muted       *sqlite.MutedChannelModel
	Cookies     *sqlite.CookieModel
	Rules       *sqlite.RuleModel
	Paths       models.ImagePaths
	Websocket   *websocket.Manager
}

func NewApp(db *sql.DB, imagePath string) *App {
	return &App{
		DB:          db,
		Users:       &sqlite.UserModel{DB: db},
		Posts:       &sqlite.PostModel{DB: db},
		Reactions:   &sqlite.ReactionModel{DB: db},
		Saved:       &sqlite.SavedModel{DB: db},
		Mods:        &sqlite.ModModel{DB: db},
		Comments:    &sqlite.CommentModel{DB: db},
		Images:      &sqlite.ImageModel{DB: db},
		Channels:    &sqlite.ChannelModel{DB: db},
		Flags:       &sqlite.FlagModel{DB: db},
		Loyalty:     &sqlite.LoyaltyModel{DB: db},
		Memberships: &sqlite.MembershipModel{DB: db},
		Muted:       &sqlite.MutedChannelModel{DB: db},
		Cookies:     &sqlite.CookieModel{DB: db},
		Rules:       &sqlite.RuleModel{DB: db},

		Paths: models.ImagePaths{
			Channel: imagePath + "channel-images/",
			Post:    imagePath + "post-images/",
			User:    imagePath + "user-images/",
		},
	}
}

func InitializeApp() (*App, func(), error) {
	cfg := initConfig()
	// Initialize DB
	initDB, err := db.InitDB(cfg.DBPath, cfg.SchemaPath)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}

	// Verify connection
	if err = initDB.Ping(); err != nil {
		initDB.Close() // Close DB if ping fails
		log.Printf(Colors.Red+"Error pinging DB: %v\n"+Colors.Reset, err)
		return nil, nil, err
	}

	// Get version
	var dbVersion string
	if err = initDB.QueryRow("select sqlite_version()").Scan(&dbVersion); err != nil {
		log.Printf(Colors.Red+"Error fetching SQLite version: %v\n"+Colors.Reset, err)
	}
	log.Printf(ErrorMsgs.DBSuccess, cfg.DBType, dbVersion)

	// App instance with DB reference
	appInstance := NewApp(initDB, cfg.ImagePath)

	// Cleanup function to close DB connection
	cleanup := func() {
		fmt.Println("Closing database connection...")
		if err := initDB.Close(); err != nil {
			log.Println(Colors.Red+"Error closing DB: %v"+Colors.Reset, err)
		} else {
			log.Println(Colors.Green + "Database closed successfully." + Colors.Reset)
		}
	}

	return appInstance, cleanup, nil
}
