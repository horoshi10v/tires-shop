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
	repo      domain.OrderRepository
	logger    *slog.Logger
	notifier  telegram.Notifier
	botSender telegram.Sender
}

func NewOrderService(
	repo domain.OrderRepository,
	logger *slog.Logger,
	notifier telegram.Notifier,
	botSender telegram.Sender,
) domain.OrderService {
	return &orderService{
		repo:      repo,
		logger:    logger,
		notifier:  notifier,
		botSender: botSender,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, dto domain.CreateOrderDTO, userID *uuid.UUID) (uuid.UUID, error) {
	s.logger.Info("processing new order", slog.String("customer", dto.CustomerName))

	orderID, err := s.repo.CreateOrderTx(ctx, dto, userID)
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

func (s *orderService) SendOrderMessage(ctx context.Context, id uuid.UUID, message string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order.CustomerTelegramID == nil || *order.CustomerTelegramID == 0 {
		return fmt.Errorf("order does not have customer_telegram_id")
	}

	telegramMessageID, err := s.botSender.SendMessage(*order.CustomerTelegramID, message)
	if err != nil {
		return err
	}

	_, err = s.repo.CreateMessage(ctx, domain.CreateOrderMessageDTO{
		OrderID:            order.ID,
		CustomerTelegramID: *order.CustomerTelegramID,
		Direction:          domain.OrderMessageDirectionOutbound,
		MessageText:        message,
		TelegramMessageID:  telegramMessageID,
	})
	return err
}

func (s *orderService) ListOrderMessages(ctx context.Context, id uuid.UUID) ([]domain.OrderMessage, error) {
	return s.repo.ListMessages(ctx, id)
}

func (s *orderService) ProcessInboundMessage(ctx context.Context, dto domain.InboundOrderMessageDTO) error {
	if dto.CustomerTelegramID == 0 || dto.TelegramMessageID == 0 || dto.MessageText == "" || dto.ReplyToTelegramMessageID == nil {
		return nil
	}

	parentMessage, err := s.repo.GetMessageByTelegramMeta(ctx, dto.CustomerTelegramID, *dto.ReplyToTelegramMessageID)
	if err != nil {
		return err
	}
	if parentMessage == nil {
		return nil
	}

	_, err = s.repo.CreateMessage(ctx, domain.CreateOrderMessageDTO{
		OrderID:                  parentMessage.OrderID,
		CustomerTelegramID:       dto.CustomerTelegramID,
		Direction:                domain.OrderMessageDirectionInbound,
		MessageText:              dto.MessageText,
		TelegramMessageID:        dto.TelegramMessageID,
		ReplyToTelegramMessageID: dto.ReplyToTelegramMessageID,
	})
	return err
}

func (s *orderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.OrderResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) ListOrders(ctx context.Context, filter domain.OrderFilter) ([]domain.OrderResponse, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	s.logger.Debug("fetching orders list", slog.Int("page", filter.Page))
	return s.repo.List(ctx, filter)
}

func (s *orderService) ListMyOrders(ctx context.Context, userID uuid.UUID, filter domain.OrderFilter) ([]domain.OrderResponse, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	s.logger.Debug("fetching user orders history", slog.String("user_id", userID.String()), slog.Int("page", filter.Page))
	return s.repo.ListByUserID(ctx, userID, filter)
}
