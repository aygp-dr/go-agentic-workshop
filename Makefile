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
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Please install it first:"; \
		echo "  - Linux/macOS: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
		echo "  - FreeBSD (pkg): pkg install golangci-lint"; \
		echo "  - FreeBSD (binary): curl -L -o golangci-lint.tar.gz https://github.com/golangci/golangci-lint/releases/download/v2.1.6/golangci-lint-2.1.6-freebsd-amd64.tar.gz && tar -xzf golangci-lint.tar.gz && mv golangci-lint-2.1.6-freebsd-amd64/golangci-lint $$(go env GOPATH)/bin/"; \
		echo "  - Windows: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./benchmarks/...