-- Migration: Add Updated column to Users table if missing
-- This handles databases created before the Updated column was added to schema
-- Context: Some WSL environments may have old databases without this column

-- The migration runner handles "duplicate column name" errors gracefully
-- If the column already exists, this migration is skipped and marked as applied

BEGIN TRANSACTION;

ALTER TABLE Users ADD COLUMN Updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;

COMMIT;
