package main

import "time"

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	UserID    string    `json:"userId" description:"Unique user identifier" example:"user-123" validate:"required,uuid4"`
	Email     string    `json:"email" description:"User email address" example:"john.doe@example.com" validate:"required,email"`
	Username  string    `json:"username" description:"User's display name" example:"johndoe" validate:"required,alphanum,min=3,max=20"`
	CreatedAt time.Time `json:"createdAt" description:"Timestamp when the user was created" validate:"required"`
}

// UserUpdatedEvent represents a user update event
type UserUpdatedEvent struct {
	UserID    string    `json:"userId" description:"Unique user identifier" example:"user-123" validate:"required,uuid4"`
	Email     string    `json:"email,omitempty" description:"Updated email address" example:"john.updated@example.com" validate:"omitempty,email"`
	Username  string    `json:"username,omitempty" description:"Updated username" example:"johndoe_updated" validate:"omitempty,alphanum,min=3,max=20"`
	UpdatedAt time.Time `json:"updatedAt" description:"Timestamp when the user was updated" validate:"required"`
}

// UserUpdateResponse represents the response to a user update
type UserUpdateResponse struct {
	Success bool   `json:"success" description:"Whether the update was successful" example:"true"`
	Message string `json:"message" description:"Response message" example:"User updated successfully"`
}

// OrderPlacedEvent represents an order placement event
type OrderPlacedEvent struct {
	OrderID    string      `json:"orderId" description:"Unique order identifier" example:"order-456" validate:"required,uuid4"`
	UserID     string      `json:"userId" description:"ID of the user who placed the order" example:"user-123" validate:"required,uuid4"`
	TotalPrice float64     `json:"totalPrice" description:"Total order price in USD" example:"99.99" validate:"required,gte=0"`
	Items      []OrderItem `json:"items" description:"List of items in the order" validate:"required,min=1,dive"`
	PlacedAt   time.Time   `json:"placedAt" description:"Timestamp when the order was placed" validate:"required"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string  `json:"productId" description:"Product identifier" example:"prod-1" validate:"required,min=1"`
	Quantity  int     `json:"quantity" description:"Number of items ordered" example:"2" validate:"required,gte=1,lte=1000"`
	Price     float64 `json:"price" description:"Price per item in USD" example:"29.99" validate:"required,gte=0"`
}

// OrderShippedEvent represents an order shipment event
type OrderShippedEvent struct {
	OrderID        string    `json:"orderId" description:"Unique order identifier" example:"order-456" validate:"required,uuid4"`
	TrackingNumber string    `json:"trackingNumber" description:"Shipping tracking number" example:"TRK123456789" validate:"required,alphanum,min=5,max=50"`
	Carrier        string    `json:"carrier" description:"Shipping carrier name" example:"UPS" validate:"required,oneof=UPS FedEx USPS DHL"`
	ShippedAt      time.Time `json:"shippedAt" description:"Timestamp when the order was shipped" validate:"required"`
}

// GetUserRequest represents a request to get user details
type GetUserRequest struct {
	UserID string `json:"userId" description:"ID of the user to retrieve" example:"user-123" validate:"required,uuid4"`
}

// GetUserResponse represents the response with user details
type GetUserResponse struct {
	UserID    string    `json:"userId" description:"Unique user identifier" example:"user-123" validate:"required,uuid4"`
	Email     string    `json:"email" description:"User email address" example:"john.doe@example.com" validate:"required,email"`
	Username  string    `json:"username" description:"User's display name" example:"johndoe" validate:"required,alphanum,min=3,max=20"`
	CreatedAt time.Time `json:"createdAt" description:"Timestamp when the user was created" validate:"required"`
	Found     bool      `json:"found" description:"Whether the user was found" example:"true" validate:"required"`
}
