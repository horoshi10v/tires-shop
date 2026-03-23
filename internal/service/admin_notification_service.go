package service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/infrastructure/telegram"
)

type adminNotificationService struct {
	repo     domain.AdminNotificationRepository
	userRepo domain.UserRepository
	sender   telegram.Sender
	logger   *slog.Logger
}

func NewAdminNotificationService(
	repo domain.AdminNotificationRepository,
	userRepo domain.UserRepository,
	sender telegram.Sender,
	logger *slog.Logger,
) domain.AdminNotificationService {
	return &adminNotificationService{
		repo:     repo,
		userRepo: userRepo,
		sender:   sender,
		logger:   logger,
	}
}

func (s *adminNotificationService) NotifyNewOrder(ctx context.Context, order *domain.OrderResponse) error {
	if order == nil {
		return nil
	}

	title := fmt.Sprintf("Нове замовлення #%s", shortOrderID(order.ID))
	body := fmt.Sprintf(
		"%s\nКлієнт: %s\nТелефон: %s\nТовари: %s\nСума: %.2f грн",
		title,
		buildCustomerLine(order),
		fallbackString(order.CustomerPhone, "Не вказано"),
		buildItemsSummary(order.Items),
		order.TotalAmount,
	)

	payload, _ := json.Marshal(map[string]any{
		"event":          "order_created",
		"order_id":       order.ID,
		"status":         order.Status,
		"items":          order.Items,
		"total_amount":   order.TotalAmount,
		"customer_name":  order.CustomerName,
		"customer_phone": order.CustomerPhone,
	})

	return s.createAndDispatch(ctx, domain.CreateAdminNotificationDTO{
		Type:               domain.AdminNotificationTypeOrderCreated,
		Title:              title,
		Body:               body,
		OrderID:            pointerToUUID(order.ID),
		CustomerName:       order.CustomerName,
		CustomerPhone:      order.CustomerPhone,
		CustomerUsername:   order.CustomerUsername,
		CustomerTelegramID: order.CustomerTelegramID,
		Payload:            payload,
	}, buildOrderCreatedTelegramBody(order, title))
}

func (s *adminNotificationService) NotifyCustomerMessage(ctx context.Context, order *domain.OrderResponse, messageText string) error {
	if order == nil || strings.TrimSpace(messageText) == "" {
		return nil
	}

	title := fmt.Sprintf("Нове повідомлення по замовленню #%s", shortOrderID(order.ID))
	body := fmt.Sprintf(
		"%s\nКлієнт: %s\nТелефон: %s\nТовари: %s\nПовідомлення: %s",
		title,
		buildCustomerLine(order),
		fallbackString(order.CustomerPhone, "Не вказано"),
		buildItemsSummary(order.Items),
		messageText,
	)

	payload, _ := json.Marshal(map[string]any{
		"event":          "customer_message",
		"order_id":       order.ID,
		"items":          order.Items,
		"message_text":   messageText,
		"customer_name":  order.CustomerName,
		"customer_phone": order.CustomerPhone,
	})

	return s.createAndDispatch(ctx, domain.CreateAdminNotificationDTO{
		Type:               domain.AdminNotificationTypeCustomerMessage,
		Title:              title,
		Body:               body,
		OrderID:            pointerToUUID(order.ID),
		CustomerName:       order.CustomerName,
		CustomerPhone:      order.CustomerPhone,
		CustomerUsername:   order.CustomerUsername,
		CustomerTelegramID: order.CustomerTelegramID,
		Payload:            payload,
	}, buildCustomerMessageTelegramBody(order, title, messageText))
}

func (s *adminNotificationService) List(ctx context.Context, filter domain.AdminNotificationFilter) ([]domain.AdminNotification, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	return s.repo.List(ctx, filter)
}

func (s *adminNotificationService) MarkRead(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkRead(ctx, id)
}

func (s *adminNotificationService) createAndDispatch(ctx context.Context, dto domain.CreateAdminNotificationDTO, telegramBody string) error {
	if _, err := s.repo.Create(ctx, dto); err != nil {
		return err
	}

	admins, _, err := s.userRepo.List(ctx, domain.UserFilter{
		Page:     1,
		PageSize: 1000,
		Role:     domain.RoleAdmin,
	})
	if err != nil {
		s.logger.Warn("failed to list admins for notifications", slog.String("error", err.Error()))
		return nil
	}

	for _, admin := range admins {
		if admin.TelegramID == 0 {
			continue
		}

		if _, err := s.sender.SendHTMLMessage(admin.TelegramID, telegramBody); err != nil {
			s.logger.Warn("failed to send admin telegram notification",
				slog.Int64("telegram_id", admin.TelegramID),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

func buildCustomerLine(order *domain.OrderResponse) string {
	parts := []string{}
	if strings.TrimSpace(order.CustomerName) != "" {
		parts = append(parts, order.CustomerName)
	}
	if strings.TrimSpace(order.CustomerUsername) != "" {
		parts = append(parts, "@"+strings.TrimPrefix(order.CustomerUsername, "@"))
	}
	if len(parts) == 0 && order.CustomerTelegramID != nil {
		parts = append(parts, fmt.Sprintf("chat_id %d", *order.CustomerTelegramID))
	}
	if len(parts) == 0 {
		return "Невідомий клієнт"
	}
	return strings.Join(parts, " • ")
}

func buildOrderCreatedTelegramBody(order *domain.OrderResponse, title string) string {
	return fmt.Sprintf(
		"<b>%s</b>\nКлієнт: %s\nТелефон: %s\nТовари: %s\nСума: %.2f грн",
		html.EscapeString(title),
		buildCustomerTelegramLink(order),
		html.EscapeString(fallbackString(order.CustomerPhone, "Не вказано")),
		html.EscapeString(buildItemsSummary(order.Items)),
		order.TotalAmount,
	)
}

func buildCustomerMessageTelegramBody(order *domain.OrderResponse, title string, messageText string) string {
	return fmt.Sprintf(
		"<b>%s</b>\nКлієнт: %s\nТелефон: %s\nТовари: %s\nПовідомлення: %s",
		html.EscapeString(title),
		buildCustomerTelegramLink(order),
		html.EscapeString(fallbackString(order.CustomerPhone, "Не вказано")),
		html.EscapeString(buildItemsSummary(order.Items)),
		html.EscapeString(messageText),
	)
}

func buildCustomerTelegramLink(order *domain.OrderResponse) string {
	displayName := strings.TrimSpace(order.CustomerName)
	if displayName == "" && strings.TrimSpace(order.CustomerUsername) != "" {
		displayName = "@" + strings.TrimPrefix(order.CustomerUsername, "@")
	}
	if displayName == "" && order.CustomerTelegramID != nil {
		displayName = fmt.Sprintf("Клієнт %d", *order.CustomerTelegramID)
	}
	if displayName == "" {
		displayName = "Невідомий клієнт"
	}

	escapedDisplayName := html.EscapeString(displayName)
	if order.CustomerTelegramID != nil && *order.CustomerTelegramID != 0 {
		return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, *order.CustomerTelegramID, escapedDisplayName)
	}

	if strings.TrimSpace(order.CustomerUsername) != "" {
		username := strings.TrimPrefix(strings.TrimSpace(order.CustomerUsername), "@")
		return fmt.Sprintf(`<a href="https://t.me/%s">%s</a>`, html.EscapeString(username), escapedDisplayName)
	}

	return escapedDisplayName
}

func buildItemsSummary(items []domain.OrderItemResponse) string {
	if len(items) == 0 {
		return "Без товарів"
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(strings.Join([]string{item.Brand, item.Model}, " "))
		if title == "" {
			title = shortOrderID(item.LotID)
		}
		parts = append(parts, fmt.Sprintf("%s x%d", title, item.Quantity))
	}
	return strings.Join(parts, ", ")
}

func fallbackString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func shortOrderID(id uuid.UUID) string {
	value := id.String()
	if len(value) > 8 {
		return value[:8]
	}
	return value
}

func pointerToUUID(id uuid.UUID) *uuid.UUID {
	return &id
}
