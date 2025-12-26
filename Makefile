.PHONY: help build build-all run test lint clean docker-build docker-up docker-down migrate deps tidy install-tools check all

# Variables
PROJECT_NAME=go-cms-service
DOCKER_COMPOSE=docker-compose
GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
GOFMT=gofmt
SERVICES=cms-admin-service cms-stats-service cms-frontend-service cms-media-service cms-crawler-service

# Build info
VERSION?=latest
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o bin/$$service ./services/$$service/cmd || exit 1; \
	done
	@echo "Build complete"

build-admin: ## Build CMS Admin Service
	@echo "Building cms-admin-service..."
	@CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o bin/cms-admin-service ./services/cms-admin-service/cmd

build-stats: ## Build CMS Stats Service
	@echo "Building cms-stats-service..."
	@CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o bin/cms-stats-service ./services/cms-stats-service/cmd

build-frontend: ## Build CMS Frontend Service
	@echo "Building cms-frontend-service..."
	@CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o bin/cms-frontend-service ./services/cms-frontend-service/cmd

build-media: ## Build CMS Media Service
	@echo "Building cms-media-service..."
	@CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o bin/cms-media-service ./services/cms-media-service/cmd

build-crawler: ## Build CMS Crawler Service
	@echo "Building cms-crawler-service..."
	@CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o bin/cms-crawler-service ./services/cms-crawler-service/cmd

run-admin: ## Run CMS Admin Service locally
	@echo "Running cms-admin-service..."
	@cd services/cms-admin-service && $(GO) run cmd/main.go

test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v -cover ./...

test-pkg: ## Run tests for pkg only
	@echo "Running tests for pkg..."
	@$(GOTEST) -v -cover ./pkg/...

test-services: ## Run tests for all services
	@echo "Running tests for services..."
	@for service in $(SERVICES); do \
		echo "Testing $$service..."; \
		$(GOTEST) -v -cover ./services/$$service/... || exit 1; \
	done

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@$(GOTEST) -v -tags=integration ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@$(GOTEST) -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@$(GOVET) ./...
	@echo "Running gofmt..."
	@$(GOFMT) -l -w .

lint-check: ## Check code formatting without modifying
	@echo "Checking code formatting..."
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "Code is not formatted. Run 'make lint' to fix."; \
		$(GOFMT) -l .; \
		exit 1; \
	fi
	@$(GOVET) ./...

fmt: ## Format code
	@echo "Formatting code..."
	@$(GOFMT) -l -w .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@$(DOCKER_COMPOSE) build

docker-up: ## Start services with Docker Compose
	@echo "Starting services..."
	@$(DOCKER_COMPOSE) up -d
	@echo "Services started"
	@echo "CMS Admin Service: http://localhost:8080"
	@echo "CMS Stats Service: http://localhost:8081"
	@echo "CMS Frontend Service: http://localhost:8082"
	@echo "CMS Media Service: http://localhost:8083"
	@echo "CMS Crawler Service: http://localhost:8084"
	@echo "MongoDB: localhost:27017"
	@echo "Redis: localhost:6379"

docker-down: ## Stop services
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) down
	@echo "Services stopped"

docker-logs: ## View Docker logs
	@$(DOCKER_COMPOSE) logs -f

docker-ps: ## List running containers
	@$(DOCKER_COMPOSE) ps

docker-clean: ## Clean Docker resources
	@echo "Cleaning Docker resources..."
	@$(DOCKER_COMPOSE) down -v --remove-orphans
	@echo "Docker cleanup complete"

migrate: ## Run database migrations
	@echo "Running migrations..."
	@cd services/cms-admin-service && $(GO) run cmd/main.go --migrate-only

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@echo "Dependencies downloaded"

tidy: ## Tidy and verify dependencies
	@echo "Tidying dependencies..."
	@$(GO) mod tidy
	@$(GO) mod verify
	@echo "Dependencies tidied"

vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	@$(GO) mod vendor
	@echo "Dependencies vendored"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

golangci-lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@golangci-lint run ./... || echo "golangci-lint not installed. Run 'make install-tools'"

check: lint-check test ## Run linters and tests

verify: lint-check test golangci-lint ## Run all verification checks

ci: deps verify ## Run CI pipeline locally

all: clean deps build test ## Clean, download deps, build, and test

.DEFAULT_GOAL := help
