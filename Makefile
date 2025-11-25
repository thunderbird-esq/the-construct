.PHONY: help build run test test-unit test-integration lint fmt vet clean install docker-build docker-run

# Variables
BINARY_NAME=matrix-mud
DOCKER_IMAGE=matrix-mud
GO=go
GOFLAGS=-v
LDFLAGS=-s -w

help: ## Display this help message
	@echo "Matrix MUD - Makefile Commands"
	@echo "================================"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Dependencies installed successfully"

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME) .
	@echo "Build complete: bin/$(BINARY_NAME)"

build-all: ## Build for all platforms
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 $(GO) build -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 $(GO) build -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GO) build -o bin/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GO) build -o bin/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Cross-platform builds complete"

run: build ## Build and run the application
	@echo "Starting $(BINARY_NAME)..."
	./bin/$(BINARY_NAME)

dev: ## Run in development mode with hot reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./tests/unit/...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GO) test -v -race ./tests/integration/...

test-coverage: test-unit ## Generate test coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f $(BINARY_NAME)
	rm -f construct
	rm -f coverage.txt coverage.html
	$(GO) clean
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):latest .
	@echo "Docker image built: $(DOCKER_IMAGE):latest"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 2323:2323 -p 8080:8080 -p 9090:9090 --rm --name $(DOCKER_IMAGE) $(DOCKER_IMAGE):latest

docker-stop: ## Stop Docker container
	docker stop $(DOCKER_IMAGE)

setup-hooks: ## Setup git hooks
	@echo "Setting up git hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/sh\nmake fmt\nmake vet' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed"

ci: lint test ## Run CI pipeline locally
	@echo "CI pipeline complete"

check: fmt vet lint test ## Run all checks (format, vet, lint, test)
	@echo "All checks passed"

.DEFAULT_GOAL := help
