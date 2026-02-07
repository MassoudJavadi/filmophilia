.PHONY: dev dev-stop build test generate wire sqlc lint clean help db db-stop

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
sqlc:
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

# Show help
help:
	@echo "Available targets:"
	@echo "  dev        - Run the API on port 8080"
	@echo "  dev-stop   - Stop the API server"
	@echo "  db         - Enter PostgreSQL shell"
	@echo "  db-stop    - Stop PostgreSQL container"
	@echo "  build      - Build the binary to bin/api"
	@echo "  test       - Run tests"
	@echo "  test-cover - Run tests with coverage"
	@echo "  generate   - Generate all (sqlc + wire)"
	@echo "  sqlc       - Generate SQLC code"
	@echo "  wire       - Generate Wire DI code"
	@echo "  lint       - Run go vet"
	@echo "  tidy       - Run go mod tidy"
	@echo "  clean      - Remove build artifacts"
