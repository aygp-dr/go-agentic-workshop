.PHONY: setup run test docker-up docker-down clean

# Default target
all: setup

# Run environment check script
setup:
	@echo "Checking environment setup..."
	@bash ./setup/check-environment.sh

# Run the main agent
run:
	@echo "Starting AI Agent..."
	go run cmd/agent/main.go

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Start Docker Compose services
docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d

# Stop Docker Compose services
docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	go clean
	rm -f go.sum

# Build the agent binary
build:
	@echo "Building agent..."
	go build -o bin/agent cmd/agent/main.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./benchmarks/...