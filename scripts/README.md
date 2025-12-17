# Scripts Directory

Utility scripts for managing the Codex application.

## Database Maintenance

### `fix-wsl-database.sh`

Diagnostic and repair script for WSL database schema issues.

**Purpose:** Fixes missing `Updated` column in Users table that can occur when migrating databases between Mac and WSL environments.

**Usage:**
```bash
# Make executable (first time only)
chmod +x scripts/fix-wsl-database.sh

# Run diagnostics and apply fix
./scripts/fix-wsl-database.sh
```

**What it does:**
1. Checks if `.env` file exists
2. Reads database path from `.env`
3. Inspects Users table structure
4. Adds `Updated` column if missing
5. Shows current table structure and migration status

**When to use:**
- Fresh clone on WSL shows "no such column: Updated" errors
- Database migrated from older version
- After pulling latest changes that include schema updates

## Docker Management

### `menu.sh`
Interactive menu for Docker operations.

### `configure.sh`
Configure Docker settings (image name, container name, port).

### `reset-config.sh`
Reset Docker configuration and choose database environment.

See main README.md for detailed Docker usage instructions.
