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
// @description A comprehensive NATS-based message service for handling user and order events
// @termsOfService https://example.com/terms
// @contact.name NATS Service Team
// @contact.email nats-support@example.com
// @contact.url https://example.com/nats-support
// @license.name Apache 2.0
// @license.url https://www.apache.org/licenses/LICENSE-2.0.html
// @tag users - User management events
// @tag orders - Order processing events
// @externalDocs.description NATS Service Documentation
// @externalDocs.url https://docs.example.com/nats-service
// @protocol nats
// @protocolVersion 2.9
// @url nats://localhost:4222
// @server.name production
// @server.title Production NATS Server
// @server.summary Production message broker
// @server.description NATS server running in production environment
// @server.tag production - Production environment
// @server.tag cloud - Cloud deployment
// @server.externalDocs.description NATS server setup guide
// @server.externalDocs.url https://docs.nats.io/running-a-nats-service/introduction
// @server.variable region enum=us-east,us-west,eu-west default=us-east description=Server region
// @server.binding nats.queue production-queue

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer nc.Close()

	log.Println("Connected to NATS server at nats://localhost:4222")

	svc := &Service{nc: nc}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.SubscribeToUserEvents(ctx)
	go svc.SubscribeToOrderEvents(ctx)
	go svc.SubscribeToGetUser(ctx)

	go func() {
		time.Sleep(2 * time.Second)
		svc.PublishUserCreated()
		time.Sleep(1 * time.Second)
		svc.PublishOrderPlaced()
		time.Sleep(1 * time.Second)
		svc.RequestGetUser("user-123")
	}()

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
// @description Publishes an event when a new user is created in the system
// @payload UserCreatedEvent
// @operation.tag users
// @operation.tag events
// @operation.externalDocs.description User Creation Flow Documentation
// @operation.externalDocs.url https://docs.example.com/user-creation
// @channel.title User Creation Channel
// @channel.description Channel for broadcasting user creation events to all subscribers
// @message.contentType application/json
// @message.title User Created Message
// @message.tag user-events
// @binding.nats.queue user-creation-queue
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
// @operation.tag orders
// @message.contentType application/json
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

// RequestGetUser sends a request to get user details and waits for a response
// @type pub
// @name user.get
// @summary Get User Request
// @description Sends a request to retrieve user details by ID and waits for response
// @payload GetUserRequest
// @response GetUserResponse
func (s *Service) RequestGetUser(userID string) (*GetUserResponse, error) {
	request := GetUserRequest{
		UserID: userID,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	log.Printf("Sending user.get request: %s", data)
	msg, err := s.nc.Request("user.get", data, 5*time.Second)
	if err != nil {
		log.Printf("Error sending user.get request: %v", err)
		return nil, err
	}

	var response GetUserResponse
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		log.Printf("Error unmarshaling user.get response: %v", err)
		return nil, err
	}

	log.Printf("Received user.get response: UserID=%s, Email=%s, Found=%v",
		response.UserID, response.Email, response.Found)
	return &response, nil
}

// SubscribeToGetUser subscribes to user.get requests and responds with user details
// @type sub
// @name user.get
// @summary Get User Handler
// @description Handles requests to retrieve user details
// @payload GetUserRequest
// @response GetUserResponse
func (s *Service) SubscribeToGetUser(ctx context.Context) {
	sub, err := s.nc.Subscribe("user.get", func(msg *nats.Msg) {
		var request GetUserRequest
		if err := json.Unmarshal(msg.Data, &request); err != nil {
			log.Printf("Error unmarshaling user.get request: %v", err)
			return
		}

		log.Printf("Received user.get request: UserID=%s", request.UserID)

		// Simulate looking up user details
		response := GetUserResponse{
			UserID:    request.UserID,
			Email:     "john.doe@example.com",
			Username:  "johndoe",
			CreatedAt: time.Now(),
			Found:     true,
		}

		if msg.Reply != "" {
			respData, _ := json.Marshal(response)
			msg.Respond(respData)
		}
	})

	if err != nil {
		log.Fatal("Failed to subscribe to user.get:", err)
	}

	<-ctx.Done()
	sub.Unsubscribe()
}
