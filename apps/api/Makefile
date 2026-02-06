.PHONY: dev build test generate wire sqlc lint clean help

# Run the API in development mode
dev:
	go run ./cmd/api

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
	@echo "  dev        - Run the API in development mode"
	@echo "  build      - Build the binary to bin/api"
	@echo "  test       - Run tests"
	@echo "  test-cover - Run tests with coverage"
	@echo "  generate   - Generate all (sqlc + wire)"
	@echo "  sqlc       - Generate SQLC code"
	@echo "  wire       - Generate Wire DI code"
	@echo "  lint       - Run go vet"
	@echo "  tidy       - Run go mod tidy"
	@echo "  clean      - Remove build artifacts"
