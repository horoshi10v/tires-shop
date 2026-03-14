package service

import (
	"context"
	"log/slog"

	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/infrastructure/googlesheets"
)

type exportService struct {
	lotRepo    domain.LotRepository
	reportRepo domain.ReportRepository
	exporter   googlesheets.Exporter
	logger     *slog.Logger
}

func NewExportService(lotRepo domain.LotRepository, reportRepo domain.ReportRepository, exporter googlesheets.Exporter, logger *slog.Logger) domain.ExportService {
	return &exportService{
		lotRepo:    lotRepo,
		reportRepo: reportRepo,
		exporter:   exporter,
		logger:     logger,
	}
}

func (s *exportService) ExportInventory(ctx context.Context) (string, error) {
	s.logger.Info("generating google sheets inventory export")

	filter := domain.LotFilter{Page: 1, PageSize: 10000}
	lots, _, err := s.lotRepo.ListInternal(ctx, filter)
	if err != nil {
		s.logger.Error("failed to fetch lots for export", slog.String("error", err.Error()))
		return "", err
	}

	var inStockLots []domain.LotInternalResponse
	for _, lot := range lots {
		if lot.CurrentQuantity > 0 && lot.Status == "ACTIVE" {
			inStockLots = append(inStockLots, lot)
		}
	}

	return s.exporter.GenerateInventoryReport(ctx, inStockLots)
}

func (s *exportService) ExportPnL(ctx context.Context) (string, error) {
	s.logger.Info("generating google sheets pnl export")

	pnl, err := s.reportRepo.GetPnL(ctx)
	if err != nil {
		s.logger.Error("failed to fetch pnl for export", slog.String("error", err.Error()))
		return "", err
	}

	return s.exporter.GeneratePnLReport(ctx, pnl)
}
