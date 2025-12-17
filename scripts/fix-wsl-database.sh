#!/bin/bash
# Script to diagnose and fix missing Updated column in WSL database

set -e

echo "=== Database Diagnostics ==="
echo ""

# Get database path from .env
if [ ! -f .env ]; then
    echo "❌ No .env file found. Create one first:"
    echo "   cp .env.example .env"
    exit 1
fi

DB_PATH=$(grep DB_PATH .env | cut -d'=' -f2)
echo "📂 Database path: $DB_PATH"
echo ""

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "⚠️  Database file doesn't exist. It will be created on first run."
    echo "   Run: make build && make run"
    exit 0
fi

# Check Users table structure
echo "🔍 Checking Users table structure..."
echo ""

COLUMN_CHECK=$(sqlite3 "$DB_PATH" "PRAGMA table_info(Users);" | grep -i "Updated" || echo "")

if [ -z "$COLUMN_CHECK" ]; then
    echo "❌ Updated column is MISSING from Users table"
    echo ""
    echo "🔧 Applying fix..."

    # Add the column
    sqlite3 "$DB_PATH" "ALTER TABLE Users ADD COLUMN Updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;" 2>&1

    if [ $? -eq 0 ]; then
        echo "✅ Updated column added successfully!"
    else
        echo "⚠️  Column may already exist (this is OK)"
    fi
else
    echo "✅ Updated column exists:"
    echo "$COLUMN_CHECK"
fi

echo ""
echo "=== Current Users table structure ==="
sqlite3 "$DB_PATH" "PRAGMA table_info(Users);"

echo ""
echo "=== Migration status ==="
sqlite3 "$DB_PATH" "SELECT name FROM migrations ORDER BY id;" 2>&1 || echo "No migrations table (this is OK for fresh databases)"

echo ""
echo "Done! You can now run the application."
