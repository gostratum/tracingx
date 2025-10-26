## Consolidated Makefile for tracingx
.PHONY: all test test-coverage coverage-html lint fmt vet tidy clean check help deps build test-unit test-integration install-tools \
	up down logs run-example bench reset-test-data dev docs release version validate-version \
	update-deps bump-patch bump-minor bump-major release-dry-run release-patch release-minor release-major health

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Test parameters
TEST_TIMEOUT=30s
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

VERSION := $(shell cat .version 2>/dev/null || echo "0.0.0")

all: test

# Default target: show help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development setup
deps: ## Download module dependencies
	go mod download
	go mod tidy

build: ## Build the project
	go build ./...

# Testing
test: test-unit test-integration ## Run unit and integration tests

test-unit: ## Run unit tests (short)
	go test -v -race -short ./...

test-integration: ## Run integration tests (requires services)
	@echo "Running integration tests..."
	@go test -v -race -tags=integration ./...

test-coverage: ## Run tests with coverage and generate HTML report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Generate HTML coverage report
coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Code quality
fmt: ## Format code
	go fmt ./...
	@command -v goimports >/dev/null 2>&1 && goimports -w . || echo "goimports not installed; run 'make install-tools'"

lint: ## Run linter (golangci-lint)
	@GOLANGCI_BIN=$(go env GOPATH)/bin/golangci-lint; \
	if [ -x "$$GOLANGCI_BIN" ]; then \
		"$$GOLANGCI_BIN" run ./...; \
	elif command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: make install-tools"; exit 1; \
	fi

vet: ## Run go vet
	go vet ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@$(GOMOD) tidy

# Clean up
clean: ## Clean up build artifacts
	go clean ./...
	rm -f coverage.out coverage.html

# Run all checks (format, vet, lint, test)
check: fmt vet lint test

# Docker services (for integration testing)
up: ## Start services for integration testing
	@echo "Starting services..."
	# Add docker compose commands for tracing services (e.g., jaeger)

down: ## Stop all services
	@echo "Stopping services..."
	# Add docker compose down commands

logs: ## Show service logs
	@echo "Showing logs..."
	# Add docker compose logs commands

# Example and demo
run-example: ## Run the basic example
	@echo "Running example..."
	# Add example run commands

# Benchmarks
bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

# Quick development cycle
dev: fmt test-unit ## Quick development cycle: format, test

# Documentation
docs: ## Generate documentation
	go doc -all .

# Install development tools
install-tools: ## Install development tools used by the project
	@echo "Installing development tools..."
	@command -v goimports >/dev/null 2>&1 || \
		(go install golang.org/x/tools/cmd/goimports@latest)
	@command -v golangci-lint >/dev/null 2>&1 || \
		(go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@echo "Tools installed (may need to add $(go env GOPATH)/bin to PATH)"

# Version management helpers
version: ## Print current version
	@echo "Current version: v$(VERSION)"

validate-version: ## Validate .version file if scripts exist
	@if [ -x ./scripts/validate-version.sh ]; then ./scripts/validate-version.sh; else echo "No validate-version script present"; fi

update-deps: ## Run update-deps script if present
	@if [ -x ./scripts/update-deps.sh ]; then ./scripts/update-deps.sh; else echo "No update-deps script present"; fi

bump-patch:
	@./scripts/bump-version.sh patch

bump-minor:
	@./scripts/bump-version.sh minor

bump-major:
	@./scripts/bump-version.sh major

# Release management (delegated to scripts if present)
release-dry-run:
	@DRY_RUN=true ./scripts/release.sh $(or $(TYPE),patch)

release-patch:
	@./scripts/release.sh patch

release-minor:
	@./scripts/release.sh minor

release-major:
	@./scripts/release.sh major

release: ## Run release script (default: patch)
	@./scripts/release.sh $(or $(TYPE),patch)