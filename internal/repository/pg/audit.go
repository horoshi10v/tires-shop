package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/repository/models"
	"gorm.io/gorm"
)

type AuditRepo struct {
	db *gorm.DB
}

type auditLogRow struct {
	models.AuditLog
	Username    string
	FirstName   string
	LastName    string
	PhoneNumber string
}

func NewAuditRepository(db *gorm.DB) domain.AuditLogRepository {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLogResponse, int64, error) {
	var rows []auditLogRow
	var total int64

	query := r.db.WithContext(ctx).
		Table("audit_logs").
		Select("audit_logs.*, users.username, users.first_name, users.last_name, users.phone_number").
		Joins("LEFT JOIN users ON users.id = audit_logs.user_id")

	if filter.Entity != "" {
		query = query.Where("audit_logs.entity = ?", filter.Entity)
	}

	if filter.Action != "" {
		query = query.Where("audit_logs.action = ?", filter.Action)
	}

	if filter.User != "" {
		search := "%" + filter.User + "%"
		query = query.Where(
			"users.username ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.phone_number ILIKE ?",
			search, search, search, search,
		)
	}

	if filter.StartDate != nil {
		query = query.Where("audit_logs.created_at >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("audit_logs.created_at < ?", filter.EndDate.AddDate(0, 0, 1))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("audit_logs.created_at DESC").Offset(offset).Limit(filter.PageSize).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch audit logs: %w", err)
	}

	response := make([]domain.AuditLogResponse, 0, len(rows))
	for _, row := range rows {
		userLabel := strings.TrimSpace(row.Username)
		if userLabel == "" {
			fullName := strings.TrimSpace(fmt.Sprintf("%s %s", row.FirstName, row.LastName))
			if fullName != "" {
				userLabel = fullName
			}
		}
		if userLabel == "" && row.PhoneNumber != "" {
			userLabel = row.PhoneNumber
		}
		if userLabel == "" {
			userLabel = row.UserID.String()
		}

		response = append(response, domain.AuditLogResponse{
			ID:        row.ID,
			Entity:    row.Entity,
			EntityID:  row.EntityID,
			Action:    row.Action,
			UserID:    row.UserID,
			UserLabel: userLabel,
			Comment:   row.Comment,
			OldValue:  json.RawMessage(row.OldValue),
			NewValue:  json.RawMessage(row.NewValue),
			CreatedAt: row.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return response, total, nil
}
