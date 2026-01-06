# Quick Start Guide

## What's Included

This NATS example demonstrates AsyncAPI code generation with:

### Files
- `main.go` - Main application with AsyncAPI annotations and NATS pub/sub logic
- `types.go` - Message type definitions (UserCreatedEvent, OrderPlacedEvent, etc.)
- `go.mod` / `go.sum` - Go module dependencies
- `Makefile` - Convenient build and run commands
- `README.md` - Complete documentation

### Message Operations

**Publishers (pub):**
- `user.created` - Publishes when a user is created
- `order.{orderId}.placed` - Publishes when an order is placed (parameterized topic)

**Subscribers (sub):**
- `user.updated` - Subscribes to user updates with request-reply pattern
- `order.{orderId}.shipped` - Subscribes to order shipment events (wildcard)

## Quick Start (3 steps)

### 1. Start NATS Server

```bash
# Using Docker (recommended)
make nats-start

# Or manually with Docker
docker run -d --name nats-local -p 4222:4222 nats:latest

# Or if you have NATS installed locally
nats-server
```

### 2. Run the Example

```bash
make run

# Or with go directly
go run .
```

You'll see output like:
```
Connected to NATS server at nats://localhost:4222
Publishing user.created event: {"userId":"user-123",...}
Publishing order.order-456.placed event: {"orderId":"order-456",...}
```

### 3. Generate AsyncAPI Spec

First, build the generator from the project root:

```bash
cd ../..
make build
cd example/nats
```

Then generate the spec:

```bash
make generate

# Or manually
../../bin/ag generate -output ./asyncapi.yaml .
```

This creates `asyncapi.yaml` with your complete API specification!

## Cleanup

```bash
# Stop NATS server
make nats-stop

# Remove generated files
make clean
```

## Next Steps

1. Modify message types in `types.go`
2. Add new publishers/subscribers in `main.go` with AsyncAPI annotations
3. Regenerate the spec with `make generate`
4. Use the spec with AsyncAPI Studio, generators, or documentation tools

## AsyncAPI Annotation Format

```go
// @type pub|sub
// @name channel.name.{parameter}
// @summary Short description
// @description Longer description
// @payload MessageTypeName
// @response ResponseTypeName  // Optional, for request-reply
func (s *Service) PublishEvent() error {
    // implementation
}
```

## Troubleshooting

**Can't connect to NATS?**
- Ensure NATS is running: `docker ps | grep nats`
- Check port 4222 is available: `lsof -i :4222`

**Build errors?**
- Run `make deps` to install dependencies
- Ensure Go 1.24+ is installed: `go version`

**Generator not found?**
- Build it from project root: `cd ../.. && make build`
