package main

import (
	"context"
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

	go func() {
		time.Sleep(2 * time.Second)
		svc.PublishUserCreated()
		time.Sleep(1 * time.Second)
		svc.PublishOrderPlaced()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
}
