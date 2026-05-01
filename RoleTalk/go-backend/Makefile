# Makefile for backforge project
# Equivalent to the provided Taskfile.yml

APP_NAME := backforge
COMPOSE := docker compose
GO := go

.PHONY: help up up-build down ps logs restart migrate-up migrate-down tidy fmt test test-integration build swagger mocks lint lint-fix fmt-ci

help:
	@task --list

# Docker Compose tasks
up:
	$(COMPOSE) up -d

up-build:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down -v

ps:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs -f --tail=200

restart:
	$(COMPOSE) restart

# Migrations via migrator
migrate-up:
	$(COMPOSE) run --rm migrator up

migrate-down:
	$(COMPOSE) run --rm migrator down

# Go tasks
tidy:
	$(GO) mod tidy

fmt:
	gofmt -w .

test:
	$(GO) test ./...

test-integration:
	$(GO) test -v -tags=integration ./test/integration/...

build:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o bin/server ./cmd/server

# Generation
swagger:
	swag init -g cmd/server/main.go -o api/swagger

mocks:
	$(GO) generate ./...

# Linting
lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

fmt-ci:
	golangci-lint fmt ./...