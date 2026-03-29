# ================================
# ✨ Geokrety Stats API Makefile ✨
# ================================

BINARY := geokrety-stats-api
CMD := ./cmd/server
GO := go
BIN_DIR := bin

.DEFAULT_GOAL := help

# Colors 🌈
GREEN  := \033[32m
YELLOW := \033[33m
CYAN   := \033[36m
RESET  := \033[0m

# --------------------------------
# 📖 Help (auto-generated)
# --------------------------------
.PHONY: help
help: ## Show this help message
	@echo ""
	@echo "$(CYAN)✨ Available targets:$(RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*##/ { printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

# --------------------------------
# 🧹 Formatting & Linting
# --------------------------------
.PHONY: fmt lint vet

fmt: ## Format Go code
	@echo "$(CYAN)→ Running gofmt...$(RESET)"
	@gofmt -w .

lint: fmt ## Run linter (go vet)
	@echo "$(CYAN)→ Running go vet...$(RESET)"
	@$(GO) vet ./...

vet: lint ## Alias for lint

# --------------------------------
# 🔨 Build & Run
# --------------------------------
.PHONY: build run clean

build: ## Build the binary
	@echo "$(CYAN)→ Building $(BINARY)...$(RESET)"
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN_DIR)/$(BINARY) $(CMD)
	@echo "$(GREEN)✔ Built $(BIN_DIR)/$(BINARY)$(RESET)"

run: ## Run the application
	@echo "$(CYAN)→ Running server...$(RESET)"
	@$(GO) run $(CMD)

clean: ## Remove built binary
	@echo "$(CYAN)→ Cleaning...$(RESET)"
	@rm -f $(BIN_DIR)/$(BINARY)
	@echo "$(GREEN)✔ Cleaned$(RESET)"

# --------------------------------
# 🧪 Testing & Modules
# --------------------------------
.PHONY: test test-integration tidy

test: ## Run tests
	@echo "$(CYAN)→ Running tests...$(RESET)"
	@$(GO) test ./...

test-integration: ## Run public API integration-style contract tests
	@echo "$(CYAN)→ Running integration tests...$(RESET)"
	@$(GO) test -tags=liveintegration ./internal/api -run '^TestPublicLiveIntegration' -count=1

tidy: ## Tidy Go modules
	@echo "$(CYAN)→ Tidying modules...$(RESET)"
	@$(GO) mod tidy

# --------------------------------
# 🐳 Docker
# --------------------------------
.PHONY: docker-build up down

docker-build: ## Build Docker image
	@echo "$(CYAN)→ Building Docker image...$(RESET)"
	@docker compose build

up: ## Start stack (force recreate)
	@echo "$(CYAN)→ Starting stack...$(RESET)"
	@docker compose up --force-recreate -d

down: ## Stop stack
	@echo "$(CYAN)→ Stopping stack...$(RESET)"
	@docker compose down
