.PHONY: help build run test clean migrate-up migrate-down docker-up docker-down lint

help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback database migrations"
	@echo "  make docker-up    - Start Docker services"
	@echo "  make docker-down  - Stop Docker services"
	@echo "  make lint         - Run linter"

build:
	@echo "Building..."
	@go build -o bin/chainfeed cmd/server/main.go

run:
	@echo "Running..."
	@go run cmd/server/main.go

test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

migrate-up:
	@echo "Running migrations..."
	@migrate -path migrations -database "postgresql://chainfeed:chainfeed@localhost:5432/chainfeed?sslmode=disable" up

migrate-down:
	@echo "Rolling back migrations..."
	@migrate -path migrations -database "postgresql://chainfeed:chainfeed@localhost:5432/chainfeed?sslmode=disable" down 1

docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

lint:
	@echo "Running linter..."
	@golangci-lint run ./...
