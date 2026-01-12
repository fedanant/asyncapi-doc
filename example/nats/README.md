# NATS AsyncAPI Example

This example demonstrates how to use the asyncapi-doc tool with a NATS-based messaging application.

## Overview

This example application showcases:
- Publishing events to NATS topics
- Subscribing to NATS topics
- Using parameterized topics (e.g., `order.{orderId}.placed`)
- Request-reply patterns with NATS
- Proper AsyncAPI annotations in Go code

## Prerequisites

1. **NATS Server**: You need a running NATS server. Install using:

```bash
# Using Docker
docker run -p 4222:4222 -p 8222:8222 nats:latest

# Or using Homebrew (macOS)
brew install nats-server
nats-server

# Or download from https://nats.io/download/
```

2. **Go**: Go 1.24 or later

3. **asyncapi-doc**: Built from the parent project

```bash
cd ../..
make build
```

## Running the Example

1. Start NATS server (in a separate terminal):

```bash
docker run -p 4222:4222 -p 8222:8222 nats:latest
```

2. Install dependencies:

```bash
cd example/nats
go mod download
```

3. Run the example application:

```bash
go run .
```

You should see output similar to:
```
Connected to NATS server at nats://localhost:4222
Publishing user.created event: {...}
Publishing order.order-456.placed event: {...}
Received user.updated event: UserID=user-123, Email=john.doe@example.com
```

## Generating AsyncAPI Specification

To generate the AsyncAPI specification from this example:

```bash
# From the project root
./bin/asyncapi-doc generate -output ./example/nats/asyncapi.yaml ./example/nats
```

This will analyze the Go code annotations and generate an AsyncAPI 2.4.0 specification file.

## AsyncAPI Annotations

The example uses these annotation tags:

### Service-level Annotations (in main function comments):
- `@title` - The API title
- `@version` - The API version
- `@protocol` - The protocol (nats, amqp, mqtt, etc.)
- `@url` - The server URL

### Operation-level Annotations (in function comments):
- `@type` - Operation type: `pub` (publish) or `sub` (subscribe)
- `@name` - Channel/topic name (supports parameters like `{orderId}`)
- `@summary` - Short summary of the operation
- `@description` - Detailed description
- `@payload` - Go type name for the message payload
- `@response` - Go type name for the response (automatically enables request-reply pattern)

## Message Types

The example includes several message types:

### User Events
- `UserCreatedEvent` - Published when a user is created
- `UserUpdatedEvent` - Received when a user is updated
- `UserUpdateResponse` - Response to user update events

### Order Events
- `OrderPlacedEvent` - Published when an order is placed
- `OrderShippedEvent` - Received when an order is shipped

## Topics

The application uses the following NATS subjects:

- `user.created` - User creation events
- `user.updated` - User update events (with reply)
- `order.{orderId}.placed` - Order placement events (parameterized)
- `order.*.shipped` - Order shipment events (wildcard subscription)

## Architecture

```
┌─────────────────┐
│   NATS Server   │
│  localhost:4222 │
└────────┬────────┘
         │
         │ nats://
         │
┌────────┴────────┐
│   Application   │
│                 │
│  Publishers:    │
│  - user.created │
│  - order.placed │
│                 │
│  Subscribers:   │
│  - user.updated │
│  - order.shipped│
└─────────────────┘
```

## Next Steps

1. Modify the message types in `types.go`
2. Add new operations in `main.go` with proper annotations
3. Re-generate the AsyncAPI spec
4. Use the generated spec with AsyncAPI tools (documentation generators, code generators, etc.)

## References

- [NATS Documentation](https://docs.nats.io/)
- [AsyncAPI Specification](https://www.asyncapi.com/docs/specifications/v2.4.0)
- [NATS Go Client](https://github.com/nats-io/nats.go)
