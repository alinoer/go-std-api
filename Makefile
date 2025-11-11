.PHONY: build run test clean migrate-up migrate-down migrate-create seed

# Build the application
build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/seeder cmd/seeder/main.go

# Run the server
run:
	go run cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run database migrations up
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

# Run database migrations down
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Seed the database
seed:
	go run cmd/seeder/main.go

# Install dependencies
deps:
	go mod tidy
	go mod download

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then cp .env.example .env; fi
	@echo "Please edit .env with your database credentials"

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run the application in development mode with auto-reload (requires air)
dev:
	air