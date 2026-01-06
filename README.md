# AsyncAPI Generator

AsyncAPI Generator is a code-to-spec tool that automatically generates AsyncAPI specifications from your Go code, similar to how [Swag](https://github.com/swaggo/swag) works for OpenAPI.

## Installation

### Using Go install (Recommended)

```bash
go install github.com/fedanant/asyncapi-generator/cmd/asyncapi-generator@latest
```

## Quick Start

### 1. Annotate your Go code with AsyncAPI comments

Add service-level annotations to your main function:

```go
// @title User Service API
// @version 1.0.0
// @protocol nats
// @url nats://localhost:4222

func main() {
    // Your application code
}
```

Add operation-level annotations to your publish/subscribe functions:

```go
// PublishUserCreated publishes a user creation event
// @type pub
// @name user.created
// @summary User Created Event
// @description Publishes an event when a new user is created in the system
// @payload UserCreatedEvent
func (s *Service) PublishUserCreated(event UserCreatedEvent) error {
    // Publish logic
    return nil
}

// SubscribeToOrders subscribes to order events
// @type sub
// @name order.{orderId}.placed
// @summary Order Placed Event
// @description Subscribes to order placement events with order ID parameter
// @payload OrderPlacedEvent
func (s *Service) SubscribeToOrders(ctx context.Context) error {
    // Subscribe logic
    return nil
}
```

Define your event types with JSON tags:

```go
type UserCreatedEvent struct {
    UserID    string    `json:"userId"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    CreatedAt time.Time `json:"createdAt"`
}

type OrderPlacedEvent struct {
    OrderID    string    `json:"orderId"`
    UserID     string    `json:"userId"`
    TotalPrice float64   `json:"totalPrice"`
    PlacedAt   time.Time `json:"placedAt"`
}
```

> **💡 Tip:** See the [AsyncAPI Annotations Reference](#asyncapi-annotations-reference) section below for complete documentation of all available annotations.

### 2. Generate AsyncAPI specification

```bash
# Or use the shorthand binary name 'ag'
ag generate -output ./asyncapi.yaml ./path/to/your/code
```

### 3. Generate HTML documentation (using AsyncAPI Generator)

Once you have your AsyncAPI specification, you can generate beautiful HTML documentation using the [AsyncAPI Generator](https://github.com/asyncapi/generator) (Node.js):

```bash
# Install AsyncAPI Generator (Node.js) globally
npm install -g @asyncapi/generator

# Generate HTML documentation from your spec
asyncapi generate fromTemplate ./asyncapi.yaml @asyncapi/html-template -o ./docs

# Open the generated documentation
open ./docs/index.html
```

**Alternative: Use AsyncAPI Studio**

You can also preview and validate your spec using [AsyncAPI Studio](https://studio.asyncapi.com):
- Visit https://studio.asyncapi.com
- Import your `asyncapi.yaml` file
- View documentation, validate the spec, and export to various formats

## AsyncAPI Annotations Reference

The generator uses special comment annotations (tags starting with `@`) to extract AsyncAPI information from your Go code.

### Service-Level Annotations

Add these annotations to your `main()` function or package-level comments:

```go
// @title NATS Message Service
// @version 1.0.0
// @protocol nats
// @url nats://localhost:4222

func main() {
    // Your application code
}
```

**Available tags:**

| Tag | Description | Required | Example |
|-----|-------------|----------|---------|
| `@title` | API title/name | Yes | `@title Order Management API` |
| `@version` | API version | Yes | `@version 1.0.0` |
| `@protocol` | Message protocol | Yes | `@protocol nats`, `@protocol amqp`, `@protocol mqtt` |
| `@url` | Server URL | Yes | `@url nats://localhost:4222` |

### Operation-Level Annotations

Add these annotations to functions that publish or subscribe to messages:

```go
// PublishUserCreated publishes a user creation event
// @type pub
// @name user.created
// @summary User Created Event
// @description Publishes an event when a new user is created
// @payload UserCreatedEvent
func (s *Service) PublishUserCreated() error {
    // Implementation
}

// SubscribeToUserUpdates subscribes to user update events
// @type sub
// @name user.updated
// @summary User Updated Event
// @description Subscribes to user update events with response
// @payload UserUpdatedEvent
// @response UserUpdateResponse
func (s *Service) SubscribeToUserUpdates(ctx context.Context) {
    // Implementation
}
```

**Available tags:**

| Tag | Description | Required | Example |
|-----|-------------|----------|---------|
| `@type` | Operation type: `pub` (publish) or `sub` (subscribe) | Yes | `@type pub` |
| `@name` | Channel/topic name (supports parameters) | Yes | `@name order.{orderId}.placed` |
| `@summary` | Short operation summary | No | `@summary Order placed event` |
| `@description` | Detailed description | No | `@description Publishes when order is placed` |
| `@payload` | Go type name for message payload | Yes | `@payload OrderPlacedEvent` |
| `@response` | Go type name for response (request-reply pattern) | No | `@response OrderResponse` |

### Parameterized Channels

Use `{paramName}` syntax for dynamic channel names:

```go
// @type pub
// @name order.{orderId}.status
// @summary Order Status Update
// @payload OrderStatusEvent
func (s *Service) PublishOrderStatus(orderID string) error {
    // Will generate parameter: orderId (string)
}
```

### NATS Subject Patterns

When using NATS as the protocol, you can leverage NATS subject hierarchies and wildcards:

#### Simple Subjects

```go
// @type pub
// @name user.created
// @summary User created event
// @payload UserCreatedEvent
func (s *Service) PublishUserCreated(event UserCreatedEvent) error {
    // Publishes to: user.created
}
```

#### Hierarchical Subjects

```go
// @type pub
// @name orders.us-east.warehouse-1.shipped
// @summary Order shipped from specific warehouse
// @payload OrderShippedEvent
func (s *Service) PublishOrderShipped(event OrderShippedEvent) error {
    // Publishes to: orders.us-east.warehouse-1.shipped
}
```

#### Parameterized Subjects (Dynamic Topics)

```go
// @type pub
// @name events.{region}.{warehouseId}.inventory
// @summary Inventory update from specific warehouse
// @payload InventoryUpdateEvent
func (s *Service) PublishInventoryUpdate(region, warehouseID string, event InventoryUpdateEvent) error {
    // Publishes to: events.us-east.warehouse-1.inventory
    // Parameters: region, warehouseId (automatically extracted from subject pattern)
}
```

#### Wildcard Subscriptions

Use `*` for single token wildcard or `>` for multi-level wildcard:

```go
// Single-level wildcard - matches any single token
// @type sub
// @name orders.*.shipped
// @summary Subscribe to shipped orders from any warehouse
// @payload OrderShippedEvent
func (s *Service) SubscribeToShippedOrders(ctx context.Context) error {
    // Subscribes to: orders.*.shipped
    // Matches: orders.warehouse-1.shipped, orders.warehouse-2.shipped
}

// Multi-level wildcard - matches any number of tokens
// @type sub
// @name events.>
// @summary Subscribe to all events
// @payload GenericEvent
func (s *Service) SubscribeToAllEvents(ctx context.Context) error {
    // Subscribes to: events.>
    // Matches: events.user.created, events.order.placed, events.us-east.inventory
}

// Combined pattern
// @type sub
// @name orders.{region}.*.status
// @summary Subscribe to order status from specific region, any warehouse
// @payload OrderStatusEvent
func (s *Service) SubscribeToRegionOrderStatus(ctx context.Context) error {
    // Subscribes to: orders.{region}.*.status
    // Matches: orders.us-east.warehouse-1.status, orders.us-west.warehouse-5.status
}
```

#### Request-Reply Pattern (NATS)

For NATS request-reply operations, use `@response` to indicate a reply is expected:

```go
// @type sub
// @name user.{userId}.get
// @summary Get user by ID
// @description Request-reply pattern: client sends request, service replies with user data
// @payload GetUserRequest
// @response GetUserResponse
func (s *Service) HandleGetUser(ctx context.Context) error {
    // Subscribes to: user.{userId}.get
    // Expects to send reply with GetUserResponse
}
```

#### Emit vs Publish

Both terms are supported and equivalent:

```go
// Using "emit" terminology (common in event-driven architectures)
// @type pub
// @name user.profile.updated
// @summary Emit user profile update event
// @payload UserProfileUpdatedEvent
func (s *Service) EmitUserProfileUpdated(event UserProfileUpdatedEvent) error {
    // "Emit" and "Publish" are semantically identical
    // Use whichever term fits your team's vocabulary
}

// Using "publish" terminology (standard messaging term)
// @type pub
// @name order.placed
// @summary Publish order placed event
// @payload OrderPlacedEvent
func (s *Service) PublishOrderPlaced(event OrderPlacedEvent) error {
    // Same as emit - both produce messages to a subject/topic
}
```

**Best Practices for NATS Subjects:**

1. **Use hierarchical naming** - `namespace.entity.action` (e.g., `orders.warehouse.shipped`)
2. **Keep subjects lowercase** - Use `user.created` not `User.Created`
3. **Use dots as separators** - Not dashes or underscores
4. **Put wildcards at the end** - `orders.*.shipped` is better than `*.orders.shipped`
5. **Document parameters clearly** - Use `{paramName}` for dynamic parts
6. **Avoid deep nesting** - 3-5 levels is usually sufficient

### Complete Example

```go
package main

import (
    "context"
    "time"
)

// @title E-Commerce Event Service
// @version 2.0.0
// @protocol amqp
// @url amqp://localhost:5672

func main() {
    svc := &EventService{}
    svc.Start()
}

type EventService struct{}

// PublishOrderPlaced publishes order placement events
// @type pub
// @name orders.{userId}.placed
// @summary Order Placed
// @description Publishes when a customer places a new order
// @payload OrderPlacedEvent
func (s *EventService) PublishOrderPlaced(userID string, order OrderPlacedEvent) error {
    // Publish logic
    return nil
}

// SubscribeToPayments subscribes to payment confirmation events
// @type sub
// @name payments.confirmed
// @summary Payment Confirmed
// @description Receives payment confirmation events
// @payload PaymentConfirmedEvent
// @response PaymentAcknowledgment
func (s *EventService) SubscribeToPayments(ctx context.Context) error {
    // Subscribe logic
    return nil
}

// Event type definitions
type OrderPlacedEvent struct {
    OrderID    string    `json:"orderId"`
    UserID     string    `json:"userId"`
    TotalPrice float64   `json:"totalPrice"`
    PlacedAt   time.Time `json:"placedAt"`
}

type PaymentConfirmedEvent struct {
    PaymentID string    `json:"paymentId"`
    OrderID   string    `json:"orderId"`
    Amount    float64   `json:"amount"`
    PaidAt    time.Time `json:"paidAt"`
}

type PaymentAcknowledgment struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}
```

### Supported Protocols

The generator supports all AsyncAPI protocols:

- **NATS** - `@protocol nats`
- **AMQP** - `@protocol amqp` (RabbitMQ)
- **MQTT** - `@protocol mqtt`
- **Kafka** - `@protocol kafka`
- **WebSocket** - `@protocol ws` or `@protocol wss`
- **HTTP** - `@protocol http` or `@protocol https`
- **Redis** - `@protocol redis`
- **STOMP** - `@protocol stomp`
- **JMS** - `@protocol jms`

### Tips

1. **Always use JSON tags** - The generator uses JSON tags to determine field names in the spec
2. **One operation per function** - Each publish/subscribe operation should have its own function
3. **Type definitions** - Define your message types in the same package or imported packages
4. **Comments are optional** - Only the `@` annotations are required; regular comments are ignored
5. **Wildcard subscriptions** - For subscribers, you can use patterns like `orders.*.placed`

## Usage

### Available Commands

```
asyncapi-generator <command> [options] [arguments]
```

#### Commands

- **generate** - Generate AsyncAPI specification from Go code
- **version** - Print version information
- **help** - Show help message

### Generate Command

```bash
asyncapi-generator generate [options] <source-directory>
```

#### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-output` | Output file path for generated spec | `./asyncapi.yaml` |
| `-template` | Template to use (yaml, json, html) | `yaml` |
| `-verbose` | Enable verbose output | `false` |

### Examples

- [NATS Example](./example/nats) - Complete example with NATS message broker
- [AsyncAPI Specification](https://www.asyncapi.com/docs/reference/specification/latest)

## Generating Documentation and Code from AsyncAPI Spec

Once you've generated your AsyncAPI specification, you can use the [AsyncAPI Generator](https://github.com/asyncapi/generator) to create:

### HTML Documentation

Beautiful, interactive HTML documentation:

```bash
# Install AsyncAPI Generator
npm install -g @asyncapi/generator

# Generate HTML docs
asyncapi generate fromTemplate ./asyncapi.yaml @asyncapi/html-template -o ./docs
```

### Markdown Documentation

Markdown documentation for your repository:

```bash
asyncapi generate fromTemplate ./asyncapi.yaml @asyncapi/markdown-template -o ./docs
```

