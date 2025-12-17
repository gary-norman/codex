#!/usr/bin/env bash

# Color variables
TEAL="\033[38;2;148;226;213m"
PEACH="\033[38;2;250;179;135m"
GREEN="\033[38;2;166;227;161m"
RED="\033[38;2;243;139;168m"
YELLOW="\033[38;2;249;226;175m"
CODEX_PINK="\033[38;2;234;79;146m"
CODEX_GREEN="\033[38;2;108;207;93m"
NC="\033[0m"

# Load existing .env if it exists
if [ -f .env ]; then
  # Read existing values
  IMAGE=$(grep '^IMAGE=' .env | cut -d'=' -f2)
  CONTAINER=$(grep '^CONTAINER=' .env | cut -d'=' -f2)
  PORT=$(grep '^PORT=' .env | cut -d'=' -f2)
fi

# Set defaults if not found
IMAGE=${IMAGE:-samuishark/codex-v1.0}
CONTAINER=${CONTAINER:-codex}
PORT=${PORT:-8888}

# Determine default database paths based on environment
# WSL detection
if grep -qi "microsoft\|WSL" /proc/version 2>/dev/null; then
  DEFAULT_DEV="./identifier.sqlite"
  DEFAULT_PROD="./production.sqlite"
else
  DEFAULT_DEV="/var/lib/db-codex/dev_forum_database.db"
  DEFAULT_PROD="/var/lib/db-codex/forum_database.db"
fi

printf "${CODEX_PINK}---------------------------------------------${NC}\n"
printf "${CODEX_GREEN}> Select Database Environment${NC}\n"
printf "${CODEX_PINK}---------------------------------------------${NC}\n"
printf "${CODEX_GREEN}Select database environment:${NC}\n"
printf "${CODEX_PINK}1)${NC} development (SQLite) - ${YELLOW}%s${NC}\n" "$DEFAULT_DEV"
printf "${CODEX_PINK}2)${NC} production (SQLite) - ${YELLOW}%s${NC}\n" "$DEFAULT_PROD"
printf "Enter selection ${CODEX_PINK}[1-2]${NC}: "

read SELECTION

case $SELECTION in
1)
  DB_ENV="dev"
  DB_PATH=$DEFAULT_DEV
  ;;
2)
  DB_ENV="prod"
  DB_PATH=$DEFAULT_PROD
  ;;
*)
  printf "${RED}✗ invalid selection, defaulting to development${NC}\n"
  DB_ENV="dev"
  DB_PATH=$DEFAULT_DEV
  ;;
esac

# Write everything to .env, preserving Docker settings
cat >.env <<EOF
DB_ENV=$DB_ENV
DB_PATH=$DB_PATH
IMAGE=$IMAGE
CONTAINER=$CONTAINER
PORT=$PORT
EOF

printf "${GREEN}✓ database configuration saved to ${CODEX_PINK}.env${NC}\n"
printf "${CODEX_PINK}---------------------------------------------${NC}\n"
printf "${CODEX_GREEN}Database environment set to: ${CODEX_PINK}%s${NC}\n" "$DB_ENV"
printf "${CODEX_GREEN}Database path: ${CODEX_PINK}%s${NC}\n" "$DB_PATH"
printf "${CODEX_PINK}---------------------------------------------${NC}\n"
