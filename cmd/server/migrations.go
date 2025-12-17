package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gary-norman/forum/internal/colors"
)

var (
	migrationColors, _ = colors.UseFlavor("Mocha")
)

// isDuplicateColumnError checks if error is due to duplicate column (safe to ignore)
func isDuplicateColumnError(err error) bool {
	return strings.Contains(err.Error(), "duplicate column name")
}

// ApplyMigrations applies multiple SQL files in order and records them in Migrations table
func runMigrations(db *sql.DB, migrationFiles []string) error {
	// Ensure migrations table exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS Migrations (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL UNIQUE,
			AppliedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	// Sort files alphabetically to run in order
	sort.Strings(migrationFiles)

	for _, file := range migrationFiles {
		var exists int
		err := db.QueryRow("SELECT COUNT(1) FROM Migrations WHERE Name = ?", file).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration record: %w", err)
		}
		if exists > 0 {
			fmt.Printf("%s⊘ Skipping already applied migration:%s %s\n", migrationColors.Yellow, migrationColors.Reset, migrationColors.CodexPink+file+migrationColors.Reset)
			continue
		}

		fmt.Printf("%s> Applying migration:%s %s\n", migrationColors.CodexPink, migrationColors.Reset, migrationColors.Teal+file+migrationColors.Reset)
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Execute migration SQL (migration files manage their own transactions)
		_, err = db.Exec(string(sqlBytes))
		if err != nil {
			// Check if it's a "duplicate column" error (safe to ignore)
			if isDuplicateColumnError(err) {
				fmt.Printf("%s⊘ Migration already applied (column exists):%s %s\n", migrationColors.Yellow, migrationColors.Reset, migrationColors.CodexPink+file+migrationColors.Reset)
				// Still record it as applied
			} else {
				return fmt.Errorf("failed to execute migration %s: %w", file, err)
			}
		}

		// Record migration in separate transaction
		_, err = db.Exec("INSERT INTO Migrations (Name, AppliedAt) VALUES (?, ?)", file, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		fmt.Printf("%s✓ Successfully applied migration:%s %s\n", migrationColors.CodexGreen, migrationColors.Reset, migrationColors.CodexPink+file+migrationColors.Reset)
	}

	return nil
}

// discoverMigrations scans the migrations directory and returns all .sql files
func discoverMigrations(migrationsDir string) ([]string, error) {
	// Use filepath.Glob to find all .sql files in the migrations directory
	pattern := filepath.Join(migrationsDir, "*.sql")
	migrations, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to discover migration files: %w", err)
	}

	// Sort alphabetically (files are named with numeric prefixes like 001_, 002_, etc.)
	sort.Strings(migrations)

	return migrations, nil
}
