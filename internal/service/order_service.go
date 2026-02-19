package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/infrastructure/telegram"
)

type orderService struct {
	repo     domain.OrderRepository
	logger   *slog.Logger
	notifier telegram.Notifier
}

func NewOrderService(repo domain.OrderRepository, logger *slog.Logger, notifier telegram.Notifier) domain.OrderService {
	return &orderService{
		repo:     repo,
		logger:   logger,
		notifier: notifier,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, dto domain.CreateOrderDTO) (uuid.UUID, error) {
	s.logger.Info("processing new order", slog.String("customer", dto.CustomerName))

	orderID, err := s.repo.CreateOrderTx(ctx, dto)
	if err != nil {
		s.logger.Error("failed to create order transaction", slog.String("error", err.Error()))
		return uuid.Nil, err
	}

	s.logger.Info("order successfully created", slog.String("order_id", orderID.String()))

	// Send a Telegram notification about the new order. This is done asynchronously to avoid blocking the main flow.
	msg := fmt.Sprintf("📦 Нове замовлення від %s!\nID: %s", dto.CustomerName, orderID.String())
	s.notifier.SendAlert(msg)

	return orderID, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID, comment string) error {
	s.logger.Info("updating order status", slog.String("order_id", id.String()), slog.String("new_status", status))

	if err := s.repo.UpdateStatus(ctx, id, status, userID, comment); err != nil {
		return err
	}

	msg := fmt.Sprintf("🔄 Статус замовлення %s змінен на: %s. Коментарій: %s", id.String(), status, comment)
	s.notifier.SendAlert(msg)

	return nil
}
