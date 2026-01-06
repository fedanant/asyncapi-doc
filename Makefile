.PHONY: build clean test fmt lint lint-install install run help

BINARY_NAME=ag
BUILD_DIR=bin
MAIN_PATH=./cmd/asyncapi-generator

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

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

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
	@rm -f coverage.out coverage.html
	@echo "Clean complete"
