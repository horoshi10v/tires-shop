package pg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/repository/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AdminNotificationRepo struct {
	db *gorm.DB
}

func NewAdminNotificationRepository(db *gorm.DB) domain.AdminNotificationRepository {
	return &AdminNotificationRepo{db: db}
}

func (r *AdminNotificationRepo) Create(ctx context.Context, dto domain.CreateAdminNotificationDTO) (*domain.AdminNotification, error) {
	record := models.AdminNotification{
		Type:               string(dto.Type),
		Title:              dto.Title,
		Body:               dto.Body,
		OrderID:            dto.OrderID,
		CustomerName:       dto.CustomerName,
		CustomerPhone:      dto.CustomerPhone,
		CustomerUsername:   dto.CustomerUsername,
		CustomerTelegramID: dto.CustomerTelegramID,
		Payload:            datatypes.JSON(dto.Payload),
		IsRead:             false,
	}

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("failed to create admin notification: %w", err)
	}

	return mapAdminNotificationModel(record), nil
}

func (r *AdminNotificationRepo) List(ctx context.Context, filter domain.AdminNotificationFilter) ([]domain.AdminNotification, int64, error) {
	var rows []models.AdminNotification
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AdminNotification{})

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.IsRead != nil {
		query = query.Where("is_read = ?", *filter.IsRead)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count admin notifications: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch admin notifications: %w", err)
	}

	result := make([]domain.AdminNotification, 0, len(rows))
	for _, row := range rows {
		result = append(result, *mapAdminNotificationModel(row))
	}

	return result, total, nil
}

func (r *AdminNotificationRepo) MarkRead(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&models.AdminNotification{}).
		Where("id = ?", id).
		Update("is_read", true).Error; err != nil {
		return fmt.Errorf("failed to mark admin notification as read: %w", err)
	}

	return nil
}

func mapAdminNotificationModel(record models.AdminNotification) *domain.AdminNotification {
	var payload json.RawMessage
	if len(record.Payload) > 0 {
		payload = json.RawMessage(record.Payload)
	}

	return &domain.AdminNotification{
		ID:                 record.ID,
		Type:               domain.AdminNotificationType(record.Type),
		Title:              record.Title,
		Body:               record.Body,
		OrderID:            record.OrderID,
		CustomerName:       record.CustomerName,
		CustomerPhone:      record.CustomerPhone,
		CustomerUsername:   record.CustomerUsername,
		CustomerTelegramID: record.CustomerTelegramID,
		Payload:            payload,
		IsRead:             record.IsRead,
		CreatedAt:          record.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
