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
	Status string `json:"status" binding:"required,oneof=NEW PREPAYMENT DONE CANCELLED"`
}

// OrderRepository handles database operations for orders, including transactions.
type OrderRepository interface {
	CreateOrderTx(ctx context.Context, dto CreateOrderDTO) (uuid.UUID, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// OrderService handles business logic for orders.
type OrderService interface {
	CreateOrder(ctx context.Context, dto CreateOrderDTO) (uuid.UUID, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error
}
