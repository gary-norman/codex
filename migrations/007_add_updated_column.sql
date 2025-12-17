-- Migration: Add Updated column to Users table if missing
-- This handles databases created before the Updated column was added to schema
-- Context: Some WSL environments may have old databases without this column

BEGIN TRANSACTION;

-- SQLite doesn't support "ADD COLUMN IF NOT EXISTS"
-- This will succeed if column is missing, fail if it exists
-- The migration runner should handle the duplicate column error gracefully

ALTER TABLE Users ADD COLUMN Updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;

COMMIT;

-- Note: If you see "duplicate column name: Updated" error, this is expected
-- and means your database already has the column (no action needed)
