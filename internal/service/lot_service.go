package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/infrastructure/qrcode"
)

// lotService implements domain.LotService.
type lotService struct {
	repo   domain.LotRepository
	logger *slog.Logger
	qrGen  qrcode.Generator
}

// NewLotService initializes the business logic layer for lots.
func NewLotService(repo domain.LotRepository, logger *slog.Logger, qrGen qrcode.Generator) domain.LotService {
	return &lotService{
		repo:   repo,
		logger: logger,
		qrGen:  qrGen,
	}
}

// CreateLot handles the business rules for creating a new lot.
func (s *lotService) CreateLot(ctx context.Context, dto domain.CreateLotDTO) (uuid.UUID, error) {
	s.logger.Debug("attempting to create new lot", slog.String("brand", dto.Brand))

	// Note: Basic validation (like price > 0) is handled by Gin binding tags in the DTO.
	// Complex business validations would go here.

	// Call the repository layer to save data
	id, err := s.repo.Create(ctx, &dto)
	if err != nil {
		s.logger.Error("failed to create lot in repository", slog.String("error", err.Error()))
		return uuid.Nil, err
	}

	s.logger.Info("lot created successfully", slog.String("lot_id", id.String()))
	return id, nil
}

// UpdateLot handles updating an existing lot.
func (s *lotService) UpdateLot(ctx context.Context, id uuid.UUID, dto domain.UpdateLotDTO) error {
	s.logger.Debug("attempting to update lot", slog.String("lot_id", id.String()))

	if err := s.repo.Update(ctx, id, &dto); err != nil {
		s.logger.Error("failed to update lot", slog.String("lot_id", id.String()), slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("lot updated successfully", slog.String("lot_id", id.String()))
	return nil
}

// DeleteLot handles soft deletion of a lot.
func (s *lotService) DeleteLot(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("attempting to delete lot", slog.String("lot_id", id.String()))

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete lot", slog.String("lot_id", id.String()), slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("lot deleted successfully", slog.String("lot_id", id.String()))
	return nil
}

func (s *lotService) ListPublicLots(ctx context.Context, filter domain.LotFilter) ([]domain.LotPublicResponse, int64, error) {
	filter = sanitizePagination(filter)
	s.logger.Debug("fetching public lots", slog.Int("page", filter.Page))
	return s.repo.ListPublic(ctx, filter)
}

func (s *lotService) ListInternalLots(ctx context.Context, filter domain.LotFilter) ([]domain.LotInternalResponse, int64, error) {
	filter = sanitizePagination(filter)
	s.logger.Debug("fetching internal lots", slog.Int("page", filter.Page))
	return s.repo.ListInternal(ctx, filter)
}

func (s *lotService) GenerateLotQR(ctx context.Context, id uuid.UUID) ([]byte, error) {
	s.logger.Info("generating qr code for lot", slog.String("lot_id", id.String()))

	dataToEncode := id.String()

	// PNG 256x256
	pngBytes, err := s.qrGen.GeneratePNG(dataToEncode, 256)
	if err != nil {
		s.logger.Error("failed to generate qr png", slog.String("error", err.Error()))
		return nil, err
	}

	return pngBytes, nil
}

// Helper function to ensure pagination is valid
func sanitizePagination(filter domain.LotFilter) domain.LotFilter {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	return filter
}
