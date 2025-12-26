.PHONY: help build run test lint clean docker-build docker-up docker-down migrate

# Variables
APP_NAME=cms-service
DOCKER_COMPOSE=docker-compose
GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
GOFMT=$(GO) fmt

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	cd services/cms-service && $(GO) build -o ../../bin/$(APP_NAME) cmd/main.go
	@echo "Build complete: bin/$(APP_NAME)"

run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	cd services/cms-service && $(GO) run cmd/main.go

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -cover ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	$(GOVET) ./...
	$(GOFMT) ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	$(DOCKER_COMPOSE) build

docker-up: ## Start services with Docker Compose
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d
	@echo "Services started"
	@echo "Article Service: http://localhost:8080"
	@echo "MongoDB: localhost:27017"
	@echo "Redis: localhost:6379"

docker-down: ## Stop services
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down
	@echo "Services stopped"

docker-logs: ## View Docker logs
	$(DOCKER_COMPOSE) logs -f

docker-ps: ## List running containers
	$(DOCKER_COMPOSE) ps

migrate: ## Run database migrations
	@echo "Running migrations..."
	cd services/cms-service && $(GO) run cmd/main.go --migrate-only

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies downloaded"

tidy: ## Tidy and verify dependencies
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	$(GO) mod verify
	@echo "Dependencies tidied"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

check: lint test ## Run linters and tests

all: clean deps build test ## Clean, download deps, build, and test

.DEFAULT_GOAL := help
