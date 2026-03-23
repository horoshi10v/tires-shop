package domain

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type AdminNotificationType string

const (
	AdminNotificationTypeOrderCreated    AdminNotificationType = "ORDER_CREATED"
	AdminNotificationTypeCustomerMessage AdminNotificationType = "CUSTOMER_MESSAGE"
)

type AdminNotification struct {
	ID                 uuid.UUID             `json:"id"`
	Type               AdminNotificationType `json:"type"`
	Title              string                `json:"title"`
	Body               string                `json:"body"`
	OrderID            *uuid.UUID            `json:"order_id,omitempty"`
	CustomerName       string                `json:"customer_name,omitempty"`
	CustomerPhone      string                `json:"customer_phone,omitempty"`
	CustomerUsername   string                `json:"customer_username,omitempty"`
	CustomerTelegramID *int64                `json:"customer_telegram_id,omitempty"`
	Payload            json.RawMessage       `json:"payload,omitempty" swaggertype:"string"`
	IsRead             bool                  `json:"is_read"`
	CreatedAt          string                `json:"created_at"`
}

type AdminNotificationFilter struct {
	Page     int
	PageSize int
	Type     AdminNotificationType
	IsRead   *bool
}

type CreateAdminNotificationDTO struct {
	Type               AdminNotificationType
	Title              string
	Body               string
	OrderID            *uuid.UUID
	CustomerName       string
	CustomerPhone      string
	CustomerUsername   string
	CustomerTelegramID *int64
	Payload            json.RawMessage `swaggertype:"string"`
}

type AdminNotificationRepository interface {
	Create(ctx context.Context, dto CreateAdminNotificationDTO) (*AdminNotification, error)
	List(ctx context.Context, filter AdminNotificationFilter) ([]AdminNotification, int64, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
}

type AdminNotificationService interface {
	NotifyNewOrder(ctx context.Context, order *OrderResponse) error
	NotifyCustomerMessage(ctx context.Context, order *OrderResponse, messageText string) error
	List(ctx context.Context, filter AdminNotificationFilter) ([]AdminNotification, int64, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
}
