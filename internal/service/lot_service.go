package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

// lotService implements domain.LotService.
type lotService struct {
	repo   domain.LotRepository
	logger *slog.Logger
}

// NewLotService initializes the business logic layer for lots.
func NewLotService(repo domain.LotRepository, logger *slog.Logger) domain.LotService {
	return &lotService{
		repo:   repo,
		logger: logger,
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
