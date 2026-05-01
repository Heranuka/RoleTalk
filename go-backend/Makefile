# Makefile for the RoleTalk (InCharacter) Backend.
# This file provides a robust set of commands for development, testing, and deployment.

# --- Variables ---
APP_NAME      := roletalk
GO_DIR        := .
BIN_DIR       := $(GO_DIR)/bin
CMD_PATH      := $(GO_DIR)/cmd/app/main.go
PROTO_DIR     := api/proto
COMPOSE       := docker-compose
GO            := go

# --- Help ---
.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- Docker Environment ---
.PHONY: up up-build down restart ps logs
up: ## Start all containers in the background
	$(COMPOSE) up -d

up-build: ## Rebuild and start all containers
	$(COMPOSE) up -d --build

down: ## Stop and remove all containers and volumes
	$(COMPOSE) down -v

restart: ## Restart all services
	$(COMPOSE) restart

ps: ## List running containers
	$(COMPOSE) ps

logs: ## Follow logs from all containers
	$(COMPOSE) logs -f --tail=200

# --- Database & Migrations ---
.PHONY: migrate-up migrate-down
migrate-up: ## Run database migrations upwards
	$(COMPOSE) run --rm migrator up

migrate-down: ## Rollback database migrations
	$(COMPOSE) run --rm migrator down

# --- Code Generation ---
.PHONY: gen-proto swagger mocks
gen-proto: ## Generate Go and gRPC code from Proto files
	@echo "Generating gRPC code..."
	protoc --go_out=$(GO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GO_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/ai/v1/ai.proto
	@echo "Generating Python gRPC code..."
	python3 -m grpc_tools.protoc -Iapi/proto --python_out=voice-assistant/src/api --grpc_python_out=voice-assistant/src/api api/proto/ai/v1/ai.proto

swagger: ## Generate Swagger (OpenAPI) documentation
	swag init -g $(CMD_PATH) -o $(GO_DIR)/api/swagger --parseDependency --parseInternal

mocks: ## Generate mocks for testing (interfaces required)
	cd $(GO_DIR) && $(GO) generate ./...

# --- Quality Assurance ---
.PHONY: lint lint-fix fmt test test-integration
lint: ## Run golangci-lint based on .golangci.yml
	cd $(GO_DIR) && golangci-lint run ./...

lint-fix: ## Run golangci-lint and apply safe fixes
	cd $(GO_DIR) && golangci-lint run --fix ./...

fmt: ## Format Go source code
	cd $(GO_DIR) && $(GO) fmt ./...

test: ## Run unit tests
	cd $(GO_DIR) && $(GO) test -v -race -cover ./...

test-integration: ## Run integration tests using the 'integration' build tag
	cd $(GO_DIR) && $(GO) test -v -tags=integration ./test/integration/...

# --- Build & Maintenance ---
.PHONY: tidy build clean
tidy: ## Clean up and verify go.mod dependencies
	cd $(GO_DIR) && $(GO) mod tidy

build: ## Build a statically linked Linux AMD64 binary
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o $(BIN_DIR)/server $(CMD_PATH)

clean: ## Remove build artifacts and temporary files
	rm -rf $(BIN_DIR)
	rm -rf $(GO_DIR)/api/swagger