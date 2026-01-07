.PHONY: build run generate clean deps nats-start nats-stop help

BINARY_NAME=nats-example
BUILD_DIR=bin

# Build the example application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the example application
run: build
	@echo "Starting NATS example application..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Generate AsyncAPI specification
generate:
	@echo "Generating AsyncAPI specification..."
	@../../bin/ag generate -output ./asyncapi.yaml .
	@echo "✓ Generated: asyncapi.yaml"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Start NATS server using Docker
nats-start:
	@echo "Starting NATS server on localhost:4222..."
	@docker run -d --name nats-example -p 4222:4222 -p 8222:8222 nats:latest
	@echo "✓ NATS server started"
	@echo "  - Client URL: nats://localhost:4222"
	@echo "  - Monitor URL: http://localhost:8222"

# Stop NATS server
nats-stop:
	@echo "Stopping NATS server..."
	@docker stop nats-example 2>/dev/null || true
	@docker rm nats-example 2>/dev/null || true
	@echo "✓ NATS server stopped"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	@rm -rf $(BUILD_DIR)
	@rm -f asyncapi.yaml
	@echo "✓ Clean complete"

# Show help
help:
	@echo "NATS AsyncAPI Example - Available Commands:"
	@echo ""
	@echo "  make build       - Build the example application"
	@echo "  make deps        - Install Go dependencies"
	@echo "  make nats-start  - Start NATS server in Docker"
	@echo "  make run         - Build and run the example application"
	@echo "  make generate    - Generate AsyncAPI specification"
	@echo "  make nats-stop   - Stop NATS server"
	@echo "  make clean       - Remove generated files and binaries"
	@echo "  make help        - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make nats-start"
	@echo "  2. make deps"
	@echo "  3. make run"
	@echo "  4. make generate"
