.PHONY: dev dev-stop build test generate wire sqlc lint clean help db db-stop migrate-build migrate-up migrate-down migrate-create migrate-version migrate-force

# Run the API in development mode on port 8080
dev:
	PORT=8080 go run ./cmd/api

# Stop the API server
dev-stop:
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "No server running on port 8080"

# Enter PostgreSQL interactive shell
db:
	docker exec -it filmophilia_db psql -U user -d filmophilia

# Stop PostgreSQL container
db-stop:
	docker stop filmophilia_db

# Build the binary
build:
	go build -o bin/api ./cmd/api

# Run tests
test:
	go test ./...

# Run tests with coverage
test-cover:
	go test -cover ./...

# Generate all (sqlc + wire)
generate: sqlc wire

# Generate SQLC code
sqlc: migrate-build
	cd $(dir $(abspath $(lastword $(MAKEFILE_LIST)))) && go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

# Generate Wire dependency injection
wire:
	cd $(dir $(abspath $(lastword $(MAKEFILE_LIST)))) && go run github.com/google/wire/cmd/wire@latest ./internal/api

# Run linter
lint:
	go vet ./...

# Tidy dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/

# ============================================================
# MIGRATIONS (using golang-migrate)
# ============================================================

# Use current directory to avoid relative path hell
PWD = $(shell pwd)
DB_DIR = $(PWD)/db
MIGRATIONS_SRC = $(DB_DIR)/migrations
MIGRATIONS_BUILD = $(DB_DIR)/.build
DATABASE_URL ?= postgres://user:password@localhost:5432/filmophilia?sslmode=disable

# Build migrations (ensure the script uses absolute paths if needed)
migrate-build:
	@mkdir -p $(MIGRATIONS_BUILD)
	@$(DB_DIR)/scripts/build-migrations.sh

# Run all pending migrations
migrate-up: migrate-build
	migrate -path $(MIGRATIONS_BUILD) -database "$(DATABASE_URL)" up

# Rollback last migration
migrate-down: migrate-build
	migrate -path $(MIGRATIONS_BUILD) -database "$(DATABASE_URL)" down 1

# Show current migration version
migrate-version: migrate-build
	migrate -path $(MIGRATIONS_BUILD) -database "$(DATABASE_URL)" version

# Force set migration version (for fixing dirty state)
migrate-force: migrate-build
ifndef version
	$(error version is required. Usage: make migrate-force version=20260101000000)
endif
	migrate -path $(MIGRATIONS_BUILD) -database "$(DATABASE_URL)" force $(version)

# Show help
help:
	@echo "Available targets:"
	@echo ""
	@echo "  Development:"
	@echo "    dev           - Run the API on port 8080"
	@echo "    dev-stop      - Stop the API server"
	@echo ""
	@echo "  Database:"
	@echo "    db            - Enter PostgreSQL shell"
	@echo "    db-stop       - Stop PostgreSQL container"
	@echo ""
	@echo "  Migrations:"
	@echo "    migrate-create name=X  - Create new migration in YYYY/MM/"
	@echo "    migrate-up             - Apply all pending migrations"
	@echo "    migrate-down           - Rollback last migration"
	@echo "    migrate-version        - Show current version"
	@echo "    migrate-build          - Build flat structure for migrate"
	@echo ""
	@echo "  Build & Test:"
	@echo "    build         - Build the binary to bin/api"
	@echo "    test          - Run tests"
	@echo "    test-cover    - Run tests with coverage"
	@echo ""
	@echo "  Code Generation:"
	@echo "    generate      - Generate all (sqlc + wire)"
	@echo "    sqlc          - Generate SQLC code"
	@echo "    wire          - Generate Wire DI code"
	@echo ""
	@echo "  Other:"
	@echo "    lint          - Run go vet"
	@echo "    tidy          - Run go mod tidy"
	@echo "    clean         - Remove build artifacts"
