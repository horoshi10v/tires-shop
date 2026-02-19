package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/horoshi10v/tires-shop/internal/domain"
)

type orderService struct {
	repo   domain.OrderRepository
	logger *slog.Logger
}

func NewOrderService(repo domain.OrderRepository, logger *slog.Logger) domain.OrderService {
	return &orderService{
		repo:   repo,
		logger: logger,
	}
}

// CreateOrder handles business logic and delegates transaction to repository.
func (s *orderService) CreateOrder(ctx context.Context, dto domain.CreateOrderDTO) (uuid.UUID, error) {
	s.logger.Info("processing new order", slog.String("customer", dto.CustomerName))

	orderID, err := s.repo.CreateOrderTx(ctx, dto)
	if err != nil {
		s.logger.Error("failed to create order transaction", slog.String("error", err.Error()))
		return uuid.Nil, err
	}

	s.logger.Info("order successfully created", slog.String("order_id", orderID.String()))

	// TODO: Later emit an event to RabbitMQ/NATS here to send a Telegram notification.

	return orderID, nil
}