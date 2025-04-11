.PHONY: all build run clean test coverage deps lint format help

# Default target
all: clean build

# Build the application
build:
	@echo "Building application..."
	go build -o pickleball-court main.go

# Run the application
run:
	@echo "Running application..."
	go run main.go

# Clean build files
clean:
	@echo "Cleaning build files..."
	rm -f pickleball-court
	rm -f *.db
	rm -f *.log

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...

# Create database tables
init-db:
	@echo "Initializing database..."
	@go run main.go -init-db

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	timestamp=`date +%Y%m%d%H%M%S`; \
	echo "Creating migration $$timestamp\_$$name.sql"
	@touch migrations/$$timestamp\_$$name.sql

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@go run main.go -migrate

# Create an admin user
create-admin:
	@echo "Creating admin user..."
	@read -p "Enter admin username: " username; \
	read -p "Enter admin email: " email; \
	read -s -p "Enter admin password: " password; \
	echo ""; \
	go run main.go -create-admin -username=$$username -email=$$email -password=$$password

# Start development server with hot reload
dev:
	@if command -v air >/dev/null; then \
		air; \
	else \
		echo "Air is not installed. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	@if command -v swag >/dev/null; then \
		swag init; \
	else \
		echo "Swag is not installed. Installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Clean and build the application"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  clean        - Clean build files"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage"
	@echo "  deps         - Install dependencies"
	@echo "  lint         - Run linter"
	@echo "  format       - Format code"
	@echo "  init-db      - Initialize database tables"
	@echo "  migrate      - Run database migrations"
	@echo "  migrate-create - Create a new migration file"
	@echo "  create-admin - Create an admin user"
	@echo "  dev          - Start development server with hot reload"
	@echo "  docs         - Generate API documentation"
	@echo "  help         - Show this help message"

# Default to help if no target is specified
.DEFAULT_GOAL := help
