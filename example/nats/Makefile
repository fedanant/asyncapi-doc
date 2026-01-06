.PHONY: run generate clean deps nats-start nats-stop help

# Run the example application
run:
	@echo "Starting NATS example application..."
	@go run .

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
	@rm -f asyncapi.yaml
	@echo "✓ Clean complete"
