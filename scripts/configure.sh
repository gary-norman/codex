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
printf "${CODEX_GREEN}> configuring Database...${NC}\n"
printf "${CODEX_PINK}---------------------------------------------${NC}\n"
printf "${CODEX_GREEN}Select database environment:${NC}\n"
printf "${CODEX_PINK}1)${NC} development (SQLite)\n"
printf "${CODEX_PINK}2)${NC} production (SQLite)\n"
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

printf "${CODEX_PINK}---------------------------------------------${NC}\n"
printf "${CODEX_GREEN}> configuring Docker build and run options...${NC}\n"
printf "${CODEX_PINK}---------------------------------------------${NC}\n"
read -rp "Enter base image name (default: samuishark/codex): " BASE_IMAGE
BASE_IMAGE=${BASE_IMAGE:-samuishark/codex}

read -rp "Enter version (e.g., 1.1 for -v1.1): " VERSION
if [ -n "$VERSION" ]; then
  IMAGE="${BASE_IMAGE}-v${VERSION}"
else
  IMAGE="${BASE_IMAGE}"
fi
printf "${GREEN}✓ Full image name: ${CODEX_PINK}%s${NC}\n" "$IMAGE"

read -rp "Enter container name (default: codex): " CONTAINER
CONTAINER=${CONTAINER:-codex}

read -rp "Enter local port number (default: 8888): " PORT
PORT=${PORT:-8888}

# Write everything fresh
cat >.env <<EOF
DB_ENV=$DB_ENV
DB_PATH=$DB_PATH
IMAGE=$IMAGE
CONTAINER=$CONTAINER
PORT=$PORT
EOF

printf "${GREEN}✓ configuration saved to ${CODEX_PINK}.env${NC}\n"
printf "${CODEX_PINK}---------------------------------------------${NC}\n"
printf "${CODEX_GREEN}Configuration Summary:${NC}\n"
printf "${CODEX_GREEN}  Database: ${CODEX_PINK}%s${NC} (${CODEX_PINK}%s${NC})\n" "$DB_ENV" "$DB_PATH"
printf "${CODEX_GREEN}  Image: ${CODEX_PINK}%s${NC}\n" "$IMAGE"
printf "${CODEX_GREEN}  Container: ${CODEX_PINK}%s${NC}\n" "$CONTAINER"
printf "${CODEX_GREEN}  Port: ${CODEX_PINK}%s${NC}\n" "$PORT"
printf "${CODEX_PINK}---------------------------------------------${NC}\n"
