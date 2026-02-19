package service

import (
	"context"
	"log/slog"

	"github.com/horoshi10v/tires-shop/internal/domain"
)

type reportService struct {
	repo   domain.ReportRepository
	logger *slog.Logger
}

func NewReportService(repo domain.ReportRepository, logger *slog.Logger) domain.ReportService {
	return &reportService{repo: repo, logger: logger}
}

// GetPnLReport fetches the financial analytics.
func (s *reportService) GetPnLReport(ctx context.Context) (*domain.PnLReport, error) {
	s.logger.Info("generating P&L report")
	return s.repo.GetPnL(ctx)
}
