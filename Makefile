.PHONY: build clean test fmt lint lint-install install run help coverage coverage-html coverage-func coverage-report

BINARY_NAME=asyncapi-doc
BUILD_DIR=bin
MAIN_PATH=./cmd/asyncapi-doc
COVERAGE_DIR=coverage

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Build complete for all platforms"

# Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(MAIN_PATH)
	@echo "Installation complete"

# Run the application
run:
	@go run $(MAIN_PATH) $(ARGS)

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -v -race ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@echo "Coverage report saved: $(COVERAGE_DIR)/coverage.out"

# Generate HTML coverage report
coverage-html: coverage
	@echo "Generating HTML coverage report..."
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "HTML coverage report generated: $(COVERAGE_DIR)/coverage.html"
	@echo "Opening in browser..."
	@which open > /dev/null && open $(COVERAGE_DIR)/coverage.html || \
	 which xdg-open > /dev/null && xdg-open $(COVERAGE_DIR)/coverage.html || \
	 echo "Please open $(COVERAGE_DIR)/coverage.html in your browser"

# Show coverage function summary
coverage-func: coverage
	@echo "Coverage by function:"
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out

# Generate comprehensive coverage report
coverage-report: coverage
	@echo ""
	@echo "=== Coverage Summary ==="
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1
	@echo ""
	@echo "=== Coverage by Package ==="
	@go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./... 2>&1 | grep -E "coverage:|ok" | grep -v "no test files"
	@echo ""
	@echo "Detailed report: $(COVERAGE_DIR)/coverage.out"
	@echo "Run 'make coverage-html' to view HTML report"

# Legacy alias for backward compatibility
test-coverage: coverage-html

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting complete"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "❌ golangci-lint not installed."; \
		echo ""; \
		echo "Install it with: make lint-install"; \
		echo "Or visit: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# Install golangci-lint
lint-install:
	@echo "Installing golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		echo "✓ golangci-lint is already installed"; \
		golangci-lint version; \
	else \
		echo "Installing golangci-lint via go install..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		echo "✓ golangci-lint installed successfully"; \
	fi

# Tidy up dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "Dependencies tidied"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f coverage.out coverage.html coverage.xml
	@echo "Clean complete"

# Run all checks (lint + test + build)
check: lint test build
	@echo "All checks passed!"

# Display help
help:
	@echo "Available targets:"
	@echo "  build           - Build the application"
	@echo "  build-all       - Build for multiple platforms"
	@echo "  install         - Install binary to GOPATH/bin"
	@echo "  run             - Run the application (use ARGS='...' for arguments)"
	@echo "  test            - Run tests"
	@echo "  test-race       - Run tests with race detector"
	@echo "  coverage        - Run tests with coverage"
	@echo "  coverage-html   - Generate HTML coverage report and open in browser"
	@echo "  coverage-func   - Show coverage by function"
	@echo "  coverage-report - Show comprehensive coverage report"
	@echo "  fmt             - Format code"
	@echo "  lint            - Run linter"
	@echo "  lint-install    - Install golangci-lint"
	@echo "  tidy            - Tidy dependencies"
	@echo "  clean           - Clean build artifacts"
	@echo "  check           - Run all checks (lint + test + build)"
	@echo "  help            - Display this help message"
