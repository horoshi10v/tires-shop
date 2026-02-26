package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type warehouseService struct {
	repo   domain.WarehouseRepository
	logger *slog.Logger
}

func NewWarehouseService(repo domain.WarehouseRepository, logger *slog.Logger) domain.WarehouseService {
	return &warehouseService{repo: repo, logger: logger}
}

func (s *warehouseService) CreateWarehouse(ctx context.Context, dto domain.CreateWarehouseDTO) (uuid.UUID, error) {
	s.logger.Info("creating warehouse", slog.String("name", dto.Name))
	return s.repo.Create(ctx, dto)
}

func (s *warehouseService) GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *warehouseService) ListWarehouses(ctx context.Context) ([]domain.Warehouse, error) {
	return s.repo.List(ctx)
}

func (s *warehouseService) UpdateWarehouse(ctx context.Context, id uuid.UUID, dto domain.UpdateWarehouseDTO) error {
	s.logger.Info("updating warehouse", slog.String("id", id.String()))
	return s.repo.Update(ctx, id, dto)
}

func (s *warehouseService) DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
	s.logger.Warn("attempting to delete warehouse", slog.String("id", id.String()))
	return s.repo.Delete(ctx, id)
}
