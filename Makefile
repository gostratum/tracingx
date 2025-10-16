.PHONY: test test-coverage lint fmt vet tidy clean

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

all: test

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race -timeout $(TEST_TIMEOUT) ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -race -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Coverage report:"
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE)

# Generate HTML coverage report
coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Run linter
lint:
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install: https://golangci-lint.run/usage/install/"; exit 1; }
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@$(GOCMD) vet ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

# Run all checks (format, vet, lint, test)
check: fmt vet lint test

# Help
help:
	@echo "Available targets:"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage"
	@echo "  coverage-html   - Generate HTML coverage report"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  vet             - Run go vet"
	@echo "  tidy            - Tidy dependencies"
	@echo "  clean           - Clean build artifacts"
	@echo "  check           - Run all checks"
	@echo "  help            - Show this help"
