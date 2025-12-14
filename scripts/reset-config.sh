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
HIGHLIGHT="\033[1;30;48;2;166;227;161m"
CODEX_HIGHLIGHT_GREEN="\033[1;30;48;2;108;207;93m"
CODEX_HIGHLIGHT_PINK="\033[38;2;20;20;20;48;2;234;79;146m"

# Determine default database paths based on environment
# WSL detection
if grep -qi "microsoft\|WSL" /proc/version 2>/dev/null; then
  DEFAULT_DEV="./identifier.sqlite"
  DEFAULT_PROD="./production.sqlite"
else
  DEFAULT_DEV="/var/lib/db-codex/dev_forum_database.db"
  DEFAULT_PROD="/var/lib/db-codex/forum_database.db"
fi

printf "${RED}> resetting configuration...${NC}\n"
rm -f .env

printf "${CODEX_GREEN}Select database environment:${NC}\n"
printf "${CODEX_PINK}1)${NC} development (SQLite)\n"
printf "${CODEX_PINK}2)${NC} production (SQLite)\n"
printf "Enter selection ${CODEX_PINK}[1-2]${NC}: "

read SELECTION

DB_PATH=""
case $SELECTION in
1)
  echo "DB_ENV=dev" >>.env
  DB_PATH=$DEFAULT_DEV
  ;;
2)
  echo "DB_ENV=prod" >>.env
  DB_PATH=$DEFAULT_PROD
  ;;
*)
  printf "${RED}✗ invalid selection, defaulting to development${NC}\n"
  echo "DB_ENV=dev" >>.env
  DB_PATH=$DEFAULT_DEV
  ;;
esac

echo "DB_PATH=$DB_PATH" >>.env
printf "${YELLOW}⚠${CODEX_GREEN} configuration reset${NC} — run ${CODEX_PINK}'make configure'${NC} to set new values\n"
printf "${GREEN}✓${CODEX_GREEN} DB_ENV set to ${CODEX_PINK}%s${NC}\n" "$DB_PATH"
