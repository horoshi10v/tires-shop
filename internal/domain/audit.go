package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLogFilter struct {
	Page      int
	PageSize  int
	Entity    string
	Action    string
	User      string
	StartDate *time.Time
	EndDate   *time.Time
}

type AuditLogResponse struct {
	ID        uuid.UUID       `json:"id"`
	Entity    string          `json:"entity"`
	EntityID  uuid.UUID       `json:"entity_id"`
	Action    string          `json:"action"`
	UserID    uuid.UUID       `json:"user_id"`
	UserLabel string          `json:"user_label"`
	Comment   string          `json:"comment"`
	OldValue  json.RawMessage `json:"old_value,omitempty"`
	NewValue  json.RawMessage `json:"new_value,omitempty"`
	CreatedAt string          `json:"created_at"`
}

type AuditLogRepository interface {
	List(ctx context.Context, filter AuditLogFilter) ([]AuditLogResponse, int64, error)
}

type AuditLogService interface {
	ListAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLogResponse, int64, error)
}
