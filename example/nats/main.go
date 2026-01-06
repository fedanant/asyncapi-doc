package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

// @title NATS Message Service
// @version 1.0.0
// @protocol nats
// @url nats://localhost:4222

func main() {
	// Connect to NATS server
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer nc.Close()

	log.Println("Connected to NATS server at nats://localhost:4222")

	// Create service instance
	svc := &Service{nc: nc}

	// Start subscribers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.SubscribeToUserEvents(ctx)
	go svc.SubscribeToOrderEvents(ctx)

	// Publish some example messages
	go func() {
		time.Sleep(2 * time.Second)
		svc.PublishUserCreated()
		time.Sleep(1 * time.Second)
		svc.PublishOrderPlaced()
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
}

type Service struct {
	nc *nats.Conn
}

// PublishUserCreated publishes a user created event
// @type pub
// @name user.created
// @summary User Created Event
// @description Publishes an event when a new user is created
// @payload UserCreatedEvent
func (s *Service) PublishUserCreated() error {
	event := UserCreatedEvent{
		UserID:    "user-123",
		Email:     "john.doe@example.com",
		Username:  "johndoe",
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	log.Printf("Publishing user.created event: %s", data)
	return s.nc.Publish("user.created", data)
}

// SubscribeToUserEvents subscribes to user events
// @type sub
// @name user.updated
// @summary User Updated Event
// @description Subscribes to events when a user is updated
// @payload UserUpdatedEvent
// @response UserUpdateResponse
func (s *Service) SubscribeToUserEvents(ctx context.Context) {
	sub, err := s.nc.Subscribe("user.updated", func(msg *nats.Msg) {
		var event UserUpdatedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshaling user.updated event: %v", err)
			return
		}

		log.Printf("Received user.updated event: UserID=%s, Email=%s", event.UserID, event.Email)

		// Send response
		response := UserUpdateResponse{
			Success: true,
			Message: "User update processed successfully",
		}

		if msg.Reply != "" {
			respData, _ := json.Marshal(response)
			msg.Respond(respData)
		}
	})

	if err != nil {
		log.Fatal("Failed to subscribe to user.updated:", err)
	}

	<-ctx.Done()
	sub.Unsubscribe()
}

// PublishOrderPlaced publishes an order placed event
// @type pub
// @name order.{orderId}.placed
// @summary Order Placed Event
// @description Publishes an event when a new order is placed
// @payload OrderPlacedEvent
func (s *Service) PublishOrderPlaced() error {
	event := OrderPlacedEvent{
		OrderID:    "order-456",
		UserID:     "user-123",
		TotalPrice: 99.99,
		Items: []OrderItem{
			{ProductID: "prod-1", Quantity: 2, Price: 29.99},
			{ProductID: "prod-2", Quantity: 1, Price: 40.01},
		},
		PlacedAt: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	subject := "order.order-456.placed"
	log.Printf("Publishing %s event: %s", subject, data)
	return s.nc.Publish(subject, data)
}

// SubscribeToOrderEvents subscribes to order events
// @type sub
// @name order.{orderId}.shipped
// @summary Order Shipped Event
// @description Subscribes to events when an order is shipped
// @payload OrderShippedEvent
func (s *Service) SubscribeToOrderEvents(ctx context.Context) {
	sub, err := s.nc.Subscribe("order.*.shipped", func(msg *nats.Msg) {
		var event OrderShippedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshaling order.shipped event: %v", err)
			return
		}

		log.Printf("Received order.shipped event: OrderID=%s, TrackingNumber=%s",
			event.OrderID, event.TrackingNumber)
	})

	if err != nil {
		log.Fatal("Failed to subscribe to order.*.shipped:", err)
	}

	<-ctx.Done()
	sub.Unsubscribe()
}
