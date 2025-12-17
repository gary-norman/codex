#!/usr/bin/env bash
# Color variables
TEAL="\033[38;2;148;226;213m"
PEACH="\033[38;2;250;179;135m"
GREEN="\033[38;2;166;227;161m"
RED="\033[38;2;243;139;168m"
YELLOW="\033[38;2;249;226;175m"
BLUE="\033[38;2;137;180;250m"
CODEX_PINK="\033[38;2;234;79;146m"
CODEX_GREEN="\033[38;2;108;207;93m"
NC="\033[0m"
HIGHLIGHT="\033[1;38;2;0;0;0;48;2;166;227;161m"
CODEX_HIGHLIGHT_GREEN="\033[1;38;2;0;0;0;48;2;108;207;93m"
CODEX_HIGHLIGHT_PINK="\033[38;2;0;0;0;48;2;234;79;146m"

# Auto-detect Nerd Font by checking if terminal can render the icon properly
if [[ -n "$TERM_PROGRAM" ]] && [[ "$TERM_PROGRAM" =~ (WezTerm|Alacritty|kitty|iTerm) ]]; then
  # Known terminals that support Nerd Fonts well
  DOCKER_ICON="${BLUE}󰡨${CODEX_GREEN}"
  COMPOSE_ICON="${TEAL}󰡨${CODEX_GREEN}"
  SCRIPTS_ICON="${NC}󰯂${CODEX_GREEN}"
  ENTER_KEY=" 󱞦 Enter "
elif fc-list 2>/dev/null | grep -qi "nerd"; then
  # Check if any Nerd Font is installed via fontconfig
  DOCKER_ICON="${BLUE}󰡨${CODEX_GREEN}"
  COMPOSE_ICON="${TEAL}󰡨${CODEX_GREEN}"
  SCRIPTS_ICON="${NC}󰯂${CODEX_GREEN}"
  ENTER_KEY=" 󱞦 Enter "
else
  # Fallback to emoji
  DOCKER_ICON="🐳"
  COMPOSE_ICON="🐙"
  SCRIPTS_ICON="📜"
  ENTER_KEY=" ⏎ Enter "
fi

# Get current time in milliseconds
get_time_ms() {
  if command -v python3 &> /dev/null; then
    python3 -c 'import time; print(int(time.time() * 1000))'
  elif command -v node &> /dev/null; then
    node -e 'console.log(Date.now())'
  else
    # Fallback to seconds
    date +%s
  fi
}

# Format duration from milliseconds
format_duration() {
  local duration_ms=$1
  local hours=$((duration_ms / 3600000))
  local minutes=$(( (duration_ms % 3600000) / 60000 ))
  local seconds=$(( (duration_ms % 60000) / 1000 ))
  local ms=$((duration_ms % 1000))

  if [ $hours -gt 0 ]; then
    printf "%dh %dm %d.%03ds" "$hours" "$minutes" "$seconds" "$ms"
  elif [ $minutes -gt 0 ]; then
    printf "%dm %d.%03ds" "$minutes" "$seconds" "$ms"
  else
    printf "%d.%03ds" "$seconds" "$ms"
  fi
}

# Read arrow keys - sets global variable $KEY
read_arrow() {
  IFS= read -rsn1 KEY < /dev/tty
  if [[ $KEY == $'\x1b' ]]; then
    read -rsn2 KEY < /dev/tty
    case $KEY in
    '[A') KEY="up" ;;
    '[B') KEY="down" ;;
    esac
  fi
}

# Show menu function
show_menu() {
  local title="$1"
  local options_var="$2"
  local descs_var="$3"

  # Use eval to access array indirectly (compatible with bash 3.2)
  eval "local options=(\"\${${options_var}[@]}\")"
  eval "local descs=(\"\${${descs_var}[@]}\")"

  local selected=1
  local max_index=$((${#options[@]} - 1))

  while true; do
    clear
    printf "${CODEX_PINK}---------------------------------------------${NC}\n"
    printf "${CODEX_GREEN}%s${NC}\n" "$title"
    printf "${CODEX_PINK}---------------------------------------------${NC}\n"

    for i in "${!options[@]}"; do
      if [ "$i" -eq 0 ]; then
        num=0
      else
        num=$i
      fi

      # Use echo -e to interpret escape sequences in option text
      if [ "$i" -eq "$selected" ]; then
        printf "${CODEX_HIGHLIGHT_GREEN}%2d) " "$num"
        echo -en "${options[$i]}"
        printf " - %s${NC}\n" "${descs[$i]}"
      else
        printf "${CODEX_PINK}%2d)${CODEX_GREEN} " "$num"
        echo -en "${options[$i]}"
        printf "${NC} - %s\n" "${descs[$i]}"
      fi
    done

    printf "${CODEX_PINK}---------------------------------------------${NC}\n"
    printf "Use ↑/↓ to navigate, Enter to select, or type number (0-%d): " "$max_index"

    read_arrow

    case "$KEY" in
    up)
      selected=$(((selected - 1 + ${#options[@]}) % ${#options[@]}))
      ;;
    down)
      selected=$(((selected + 1) % ${#options[@]}))
      ;;
    '')
      MENU_CHOICE="${options[$selected]}"
      return 0
      ;;
    q | Q)
      MENU_CHOICE="exit"
      return 0
      ;;
    [0-9])
      if [ "$KEY" -ge 0 ] && [ "$KEY" -le "$max_index" ]; then
        MENU_CHOICE="${options[$KEY]}"
        return 0
      else
        printf "\n${RED}⚠ invalid number (must be 0-%d)${NC}\n" "$max_index"
        sleep 1.5
      fi
      ;;
    *)
      ;;
    esac
  done
}

# Docker submenu
docker_menu() {
  local docker_options=("back" "configure" "reset-config" "build-image" "push-image" "run-container" "stop-container")
  local docker_descs=("return to main menu" "configure Docker options" "reset configuration" "build Docker image" "push image to registry" "run Docker container" "stop and remove container")

  while true; do
    show_menu "Docker Commands" docker_options docker_descs

    case "$MENU_CHOICE" in
      back|exit)
        return
        ;;
      *)
        clear
        printf "${TEAL}Running target: ${CODEX_PINK}%s${NC}\n\n" "$MENU_CHOICE"

        START=$(get_time_ms)
        make --no-print-directory "$MENU_CHOICE"
        EXIT_CODE=$?
        END=$(get_time_ms)
        DURATION=$((END - START))

        printf "\n${CODEX_PINK}---------------------------------------------${NC}\n"
        if [ $EXIT_CODE -eq 0 ]; then
          printf "${GREEN}✓ Task completed${NC} in ${CODEX_PINK}%s${NC}\n" "$(format_duration $DURATION)"
        else
          printf "${RED}✗ Task failed${NC} (exit code: %d) after ${CODEX_PINK}%s${NC}\n" "$EXIT_CODE" "$(format_duration $DURATION)"
        fi
        printf "${CODEX_PINK}---------------------------------------------${NC}\n"

        printf "\nPress ${CODEX_HIGHLIGHT_PINK}${ENTER_KEY}${NC} to return to menu..."
        read -r dummy
        ;;
    esac
  done
}

# Docker Compose submenu
compose_menu() {
  local compose_options=("back" "compose-up" "compose-down" "compose-restart" "compose-logs" "compose-ps" "compose-build" "compose-edit")
  local compose_descs=("return to main menu" "start services (detached)" "stop and remove services" "restart services" "view service logs" "show service status" "build services" "edit docker-compose.yml")

  while true; do
    show_menu "Docker Compose Commands" compose_options compose_descs

    case "$MENU_CHOICE" in
      back|exit)
        return
        ;;
      *)
        clear
        printf "${TEAL}Running target: ${CODEX_PINK}%s${NC}\n\n" "$MENU_CHOICE"

        START=$(get_time_ms)
        make --no-print-directory "$MENU_CHOICE"
        EXIT_CODE=$?
        END=$(get_time_ms)
        DURATION=$((END - START))

        printf "\n${CODEX_PINK}---------------------------------------------${NC}\n"
        if [ $EXIT_CODE -eq 0 ]; then
          printf "${GREEN}✓ Task completed${NC} in ${CODEX_PINK}%s${NC}\n" "$(format_duration $DURATION)"
        else
          printf "${RED}✗ Task failed${NC} (exit code: %d) after ${CODEX_PINK}%s${NC}\n" "$EXIT_CODE" "$(format_duration $DURATION)"
        fi
        printf "${CODEX_PINK}---------------------------------------------${NC}\n"

        printf "\nPress ${CODEX_HIGHLIGHT_PINK}${ENTER_KEY}${NC} to return to menu..."
        read -r dummy
        ;;
    esac
  done
}

# Scripts submenu
scripts_menu() {
  local scripts_options=("back" "install-scripts" "verify-scripts" "backup-scripts")
  local scripts_descs=("return to main menu" "install/update scripts" "verify checksums" "backup scripts")

  while true; do
    show_menu "Script Management" scripts_options scripts_descs

    case "$MENU_CHOICE" in
      back|exit)
        return
        ;;
      *)
        clear
        printf "${TEAL}Running target: ${CODEX_PINK}%s${NC}\n\n" "$MENU_CHOICE"

        START=$(get_time_ms)
        make --no-print-directory "$MENU_CHOICE"
        EXIT_CODE=$?
        END=$(get_time_ms)
        DURATION=$((END - START))

        printf "\n${CODEX_PINK}---------------------------------------------${NC}\n"
        if [ $EXIT_CODE -eq 0 ]; then
          printf "${GREEN}✓ Task completed${NC} in ${CODEX_PINK}%s${NC}\n" "$(format_duration $DURATION)"
        else
          printf "${RED}✗ Task failed${NC} (exit code: %d) after ${CODEX_PINK}%s${NC}\n" "$EXIT_CODE" "$(format_duration $DURATION)"
        fi
        printf "${CODEX_PINK}---------------------------------------------${NC}\n"

        printf "\nPress ${CODEX_HIGHLIGHT_PINK}${ENTER_KEY}${NC} to return to menu..."
        read -r dummy
        ;;
    esac
  done
}

# Main menu loop
while true; do
  # Build main menu with proper color interpolation
  DOCKER_LABEL="${DOCKER_ICON} Docker"
  COMPOSE_LABEL="${COMPOSE_ICON} Compose"
  SCRIPTS_LABEL="${SCRIPTS_ICON} Scripts"
  main_options=("exit" "build" "run" "build-run" "migrate" "seed" "select-db" "$DOCKER_LABEL" "$COMPOSE_LABEL" "$SCRIPTS_LABEL")
  main_descs=("quit this menu" "build the application" "run the application" "build and run the application" "run database migrations" "seed database with initial data" "select database environment" "Docker management" "docker-compose management" "script management")

  show_menu "make commands for <codex>" main_options main_descs

  case "$MENU_CHOICE" in
    exit)
      printf "\n${TEAL}Exiting menu...${NC}\n"
      exit 0
      ;;
    "$DOCKER_LABEL")
      docker_menu
      ;;
    "$COMPOSE_LABEL")
      compose_menu
      ;;
    "$SCRIPTS_LABEL")
      scripts_menu
      ;;
    *)
      clear
      printf "${TEAL}Running target: ${CODEX_PINK}%s${NC}\n\n" "$MENU_CHOICE"

      START=$(get_time_ms)
      make --no-print-directory "$MENU_CHOICE"
      EXIT_CODE=$?
      END=$(get_time_ms)
      DURATION=$((END - START))

      printf "\n${CODEX_PINK}---------------------------------------------${NC}\n"
      if [ $EXIT_CODE -eq 0 ]; then
        printf "${GREEN}✓ Task completed${NC} in ${CODEX_PINK}%s${NC}\n" "$(format_duration $DURATION)"
      else
        printf "${RED}✗ Task failed${NC} (exit code: %d) after ${CODEX_PINK}%s${NC}\n" "$EXIT_CODE" "$(format_duration $DURATION)"
      fi
      printf "${CODEX_PINK}---------------------------------------------${NC}\n"

      printf "\nPress ${CODEX_HIGHLIGHT_PINK}${ENTER_KEY}${NC} to return to menu..."
      read -r dummy
      ;;
  esac
done
