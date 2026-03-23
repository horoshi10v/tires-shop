package service

import (
	"context"
	"log/slog"

	"github.com/horoshi10v/tires-shop/internal/domain"
)

type auditService struct {
	repo   domain.AuditLogRepository
	logger *slog.Logger
}

func NewAuditService(repo domain.AuditLogRepository, logger *slog.Logger) domain.AuditLogService {
	return &auditService{
		repo:   repo,
		logger: logger,
	}
}

func (s *auditService) ListAuditLogs(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLogResponse, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	s.logger.Debug("fetching audit logs", slog.Int("page", filter.Page), slog.String("entity", filter.Entity), slog.String("action", filter.Action))
	return s.repo.List(ctx, filter)
}
