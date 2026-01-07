.PHONY: help build run test test-coverage lint fmt clean docker-build docker-run deps

# Variables
BINARY_NAME=file-meta
DOCKER_IMAGE=file-meta
DOCKER_TAG=latest
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')
COVERAGE_FILE=coverage.out

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod verify

build: ## Build the application
	@echo "🔨 Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) -v .
	@echo "✅ Build complete: ./$(BINARY_NAME)"

run: ## Run the application locally
	@echo "🚀 Starting server..."
	go run main.go

test: ## Run all tests
	@echo "🧪 Running tests..."
	go test -v -race ./...

test-coverage: ## Run tests with coverage report
	@echo "📊 Running tests with coverage..."
	go test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "🔍 Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
		exit 1; \
	fi

fmt: ## Format code
	@echo "📝 Formatting code..."
	gofmt -s -w $(GO_FILES)
	go mod tidy
	@echo "✅ Code formatted"

clean: ## Remove build artifacts
	@echo "🧹 Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_FILE)
	rm -f coverage.html
	go clean -cache -testcache
	@echo "✅ Clean complete"

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 --env-file .env --rm $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up: ## Start services with docker-compose
	@echo "🐳 Starting docker-compose..."
	docker-compose up -d
	@echo "✅ Services started"

docker-compose-down: ## Stop docker-compose services
	@echo "🐳 Stopping docker-compose..."
	docker-compose down
	@echo "✅ Services stopped"

install: build ## Install binary to GOPATH/bin
	@echo "📦 Installing to $(GOPATH)/bin..."
	go install .
	@echo "✅ Installed"

dev: ## Run with hot reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "⚠️  air not installed. Install with:"; \
		echo "   go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

.DEFAULT_GOAL := help
