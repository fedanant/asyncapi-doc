package main

import "time"

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	UserID    string    `json:"userId"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserUpdatedEvent represents a user update event
type UserUpdatedEvent struct {
	UserID    string    `json:"userId"`
	Email     string    `json:"email,omitempty"`
	Username  string    `json:"username,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UserUpdateResponse represents the response to a user update
type UserUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// OrderPlacedEvent represents an order placement event
type OrderPlacedEvent struct {
	OrderID    string      `json:"orderId"`
	UserID     string      `json:"userId"`
	TotalPrice float64     `json:"totalPrice"`
	Items      []OrderItem `json:"items"`
	PlacedAt   time.Time   `json:"placedAt"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// OrderShippedEvent represents an order shipment event
type OrderShippedEvent struct {
	OrderID        string    `json:"orderId"`
	TrackingNumber string    `json:"trackingNumber"`
	Carrier        string    `json:"carrier"`
	ShippedAt      time.Time `json:"shippedAt"`
}
