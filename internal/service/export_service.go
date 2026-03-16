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

func (s *exportService) ExportInventory(ctx context.Context, filter domain.LotFilter) (string, error) {
	s.logger.Info("generating google sheets inventory export")

	// Ensure we get all relevant records for the report, but respect filters
	if filter.PageSize == 0 {
		filter.PageSize = 10000 // Large limit for export
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	lots, _, err := s.lotRepo.ListInternal(ctx, filter)
	if err != nil {
		s.logger.Error("failed to fetch lots for export", slog.String("error", err.Error()))
		return "", err
	}

	// Filter in-stock lots if not already filtered
	var exportLots []domain.LotInternalResponse
	for _, lot := range lots {
		// Exports usually only care about active stock, unless status filter overrides
		if lot.Status == "ACTIVE" {
			exportLots = append(exportLots, lot)
		} else if filter.Status != "" && lot.Status == filter.Status {
			// If user explicitly asked for ARCHIVED, include them
			exportLots = append(exportLots, lot)
		}
	}

	return s.exporter.GenerateInventoryReport(ctx, exportLots)
}

func (s *exportService) ExportPnL(ctx context.Context, filter domain.ReportFilter) (string, error) {
	s.logger.Info("generating google sheets pnl export")

	pnl, err := s.reportRepo.GetPnL(ctx, filter)
	if err != nil {
		s.logger.Error("failed to fetch pnl for export", slog.String("error", err.Error()))
		return "", err
	}

	return s.exporter.GeneratePnLReport(ctx, pnl)
}
