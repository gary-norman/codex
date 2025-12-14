TEAL=\033[38;2;148;226;213m
PEACH=\033[38;2;250;179;135m
GREEN=\033[38;2;166;227;161m
RED=\033[38;2;243;139;168m
CODEX_PINK=\033[38;2;234;79;146m
CODEX_GREEN=\033[38;2;108;207;93m
NC=\033[0m# No Color
HIGHLIGHT=\033[1;30;48;2;166;227;161m# bold black text on green background
CODEX_HIGHLIGHT_GREEN=\033[1;30;48;2;108;207;93m# bold black text on codex green background
CODEX_HIGHLIGHT_PINK=\033[1;30;48;2;108;207;93m# bold black text on codex green background

NOWMS = go run tools/nowms.go

# Load saved values if .env exists
-include .env

menu:
	@./scripts/menu.sh

build: ## build the web server application
	@echo "$(CODEX_PINK)> building web server application...$(NC)"
	@START=$$($(NOWMS)); \
		go build -o bin/codex github.com/gary-norman/forum/cmd/server; \
		STOP=$$($(NOWMS)); \
		DIFF=$$((STOP - START)); \
		SEC=$$((DIFF / 1000)); \
		MS=$$((DIFF % 1000)); \
		echo "$(GREEN)✓ build complete!$(NC) in $${SEC}.$${MS}s"

run: ## run the web server application
	@echo "$(CODEX_PINK)> running web server application...$(NC)"
	@START=$$($(NOWMS)); \
		bin/codex; \
		STOP=$$($(NOWMS)); \
		DIFF=$$((STOP - START)); \
		HR=$$((DIFF / 3600000)); \
		MIN=$$(((DIFF / 60000) % 60)); \
		SEC=$$(((DIFF / 1000) % 60)); \
		MS=$$((DIFF % 1000)); \
		echo "$(GREEN)✓ server stopped$(NC) Uptime: $(CODEX_PINK)$${HR}h:$${MIN}m:$${SEC}.$${MS}s"
		printf "$(CODEX_PINK)---------------------------------------------$(NC)\n"; \

build-run: ## build and run the web server application
	@$(MAKE) --no-print-directory build && $(MAKE) --no-print-directory run

migrate: ## run database migrations
	@echo "$(CODEX_PINK)> running database migrations...$(NC)"
	@START=$$($(NOWMS)); \
		bin/codex migrate; \
		STOP=$$($(NOWMS)); \
		DIFF=$$((STOP - START)); \
		SEC=$$((DIFF / 1000)); \
		MS=$$((DIFF % 1000)); \
		echo "$(GREEN)✓ migrations complete!$(NC) in $${SEC}.$${MS}s"

seed: ## seed the database with initial data
	@echo "$(CODEX_PINK)> seeding database...$(NC)"
	@START=$$($(NOWMS)); \
		bin/codex seed; \
		STOP=$$($(NOWMS)); \
		DIFF=$$((STOP - START)); \
		SEC=$$((DIFF / 1000)); \
		MS=$$((DIFF % 1000)); \
		echo "$(GREEN)✓ seeding complete!$(NC) in $${SEC}.$${MS}s"

configure: ## configure Docker build and run options
	@./scripts/configure.sh

reset-config: ## reset Docker configuration and choose db (dev/prod)
	@./scripts/reset-config.sh

build-image: ## build the Docker image
	@bash -c '\
		START=$$($(NOWMS)); \
		printf "$(CODEX_GREEN)> building Docker image $(NC)with tag: $(CODEX_PINK)%s$(NC)\n" "$(IMAGE)"; \
		docker image build -t $(IMAGE) .; \
		STOP=$$($(NOWMS)); \
		DIFF=$$((STOP - START)); \
		SEC=$$((DIFF / 1000)); \
		MS=$$((DIFF % 1000)); \
		printf "$(GREEN)✓${CODEX_GREEN} build complete!$(NC) in $(CODEX_PINK)%s.%03d$(NC)s\n" "$$SEC" "$$MS" \
	'

run-container: ## run the Docker container
	@bash -c '\
		START=$$($(NOWMS)); \
		printf "$(TEAL)> running Docker container $(NC)as $(CODEX_PINK)%s $(NC)from image: $(CODEX_PINK)%s $(NC)using port: $(CODEX_PINK)%s$(NC)\n" "$(CONTAINER)" "$(IMAGE)" "$(PORT)"; \
		docker run -d -p $(PORT):8888 --name $(CONTAINER) $(IMAGE); \
		STOP=$$($(NOWMS)); \
		DIFF=$$((STOP - START)); \
		SEC=$$((DIFF / 1000)); \
		MS=$$((DIFF % 1000)); \
		printf "$(GREEN)✓${CODEX_GREEN} task complete!$(NC)in $(CODEX_PINK)%s.%03d$(NC)s\n" "$$SEC" "$$MS" \
	'

install-scripts: ## install/update all scripts with correct permissions
	@echo "$(CODEX_PINK)> installing scripts...$(NC)"
	@chmod +x scripts/*.sh
	@echo "$(GREEN)✓ scripts are executable$(NC)"
	@echo "$(CODEX_GREEN)Scripts installed:$(NC)"
	@ls -lh scripts/*.sh | awk '{printf "  $(CODEX_PINK)%s$(NC) - %s\n", $$9, $$5}'

verify-scripts: ## verify script checksums to detect changes
	@echo "$(CODEX_PINK)> verifying scripts...$(NC)"
	@cd scripts && sha256sum *.sh > /tmp/script-checksums.txt
	@echo "$(GREEN)✓ checksums saved to /tmp/script-checksums.txt$(NC)"
	@cat /tmp/script-checksums.txt

backup-scripts: ## backup current scripts before updating
	@echo "$(CODEX_PINK)> backing up scripts...$(NC)"
	@mkdir -p scripts/backups
	@TIMESTAMP=$$(date +%Y%m%d_%H%M%S); \
		tar -czf scripts/backups/scripts_$$TIMESTAMP.tar.gz scripts/*.sh; \
		echo "$(GREEN)✓ backup created: $(CODEX_PINK)scripts/backups/scripts_$$TIMESTAMP.tar.gz$(NC)"

fix-line-endings: ## fix Windows (CRLF) line endings to Unix (LF) - for WSL users
	@echo "$(CODEX_PINK)> fixing line endings...$(NC)"
	@if command -v dos2unix >/dev/null 2>&1; then \
		dos2unix Makefile scripts/*.sh 2>/dev/null || true; \
		echo "$(GREEN)✓ line endings fixed$(NC)"; \
	else \
		echo "$(YELLOW)⚠ dos2unix not installed$(NC)"; \
		echo "$(CODEX_GREEN)Install with:$(NC) $(CODEX_PINK)sudo apt install dos2unix$(NC)"; \
		echo "$(CODEX_GREEN)Or use sed:$(NC) $(CODEX_PINK)sed -i 's/\r$$//' Makefile scripts/*.sh$(NC)"; \
	fi
