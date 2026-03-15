package domain

import (
	"context"

	"github.com/google/uuid"
)

// OrderItemDTO represents a single lot in the order request.
type OrderItemDTO struct {
	LotID    uuid.UUID `json:"lot_id" binding:"required"`
	Quantity int       `json:"quantity" binding:"required,gt=0"`
}

// CreateOrderDTO contains data required to create a new order.
type CreateOrderDTO struct {
	CustomerName  string         `json:"customer_name" binding:"required"`
	CustomerPhone string         `json:"customer_phone" binding:"required"`
	Items         []OrderItemDTO `json:"items" binding:"required,min=1"`
}

// UpdateOrderStatusDTO represents the request to change an order's status.
type UpdateOrderStatusDTO struct {
	Status  string `json:"status" binding:"required,oneof=NEW PREPAYMENT DONE CANCELLED"`
	Comment string `json:"comment"`
}

// OrderFilter defines criteria for searching orders.
type OrderFilter struct {
	Page     int
	PageSize int
	Status   string
	Customer string // Search by name or phone
}

// OrderResponse represents the order data returned to the client.
type OrderResponse struct {
	ID            uuid.UUID           `json:"id"`
	CustomerName  string              `json:"customer_name"`
	CustomerPhone string              `json:"customer_phone"`
	Status        string              `json:"status"`
	TotalAmount   float64             `json:"total_amount"`
	CreatedAt     string              `json:"created_at"`
	Items         []OrderItemResponse `json:"items"`
}

// OrderItemResponse represents a single item in the order response.
type OrderItemResponse struct {
	LotID    uuid.UUID `json:"lot_id"`
	Quantity int       `json:"quantity"`
	Price    float64   `json:"price"`
	Total    float64   `json:"total"`
}

// OrderRepository handles database operations for orders, including transactions.
type OrderRepository interface {
	CreateOrderTx(ctx context.Context, dto CreateOrderDTO, userID *uuid.UUID) (uuid.UUID, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID, comment string) error
	List(ctx context.Context, filter OrderFilter) ([]OrderResponse, int64, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, filter OrderFilter) ([]OrderResponse, int64, error)
}

// OrderService handles business logic for orders.
type OrderService interface {
	CreateOrder(ctx context.Context, dto CreateOrderDTO, userID *uuid.UUID) (uuid.UUID, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID, comment string) error
	ListOrders(ctx context.Context, filter OrderFilter) ([]OrderResponse, int64, error)
	ListMyOrders(ctx context.Context, userID uuid.UUID, filter OrderFilter) ([]OrderResponse, int64, error)
}
