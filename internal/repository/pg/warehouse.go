package pg

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/horoshi10v/tires-shop/internal/domain"
	"github.com/horoshi10v/tires-shop/internal/repository/models"
)

type WarehouseRepo struct {
	db *gorm.DB
}

func NewWarehouseRepository(db *gorm.DB) domain.WarehouseRepository {
	return &WarehouseRepo{db: db}
}

func (r *WarehouseRepo) Create(ctx context.Context, dto domain.CreateWarehouseDTO) (uuid.UUID, error) {
	dbModel := models.Warehouse{
		Name:     dto.Name,
		Location: dto.Location,
		IsActive: true,
	}

	if err := r.db.WithContext(ctx).Create(&dbModel).Error; err != nil {
		return uuid.Nil, fmt.Errorf("failed to create warehouse: %w", err)
	}

	return dbModel.ID, nil
}

func (r *WarehouseRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	var dbModel models.Warehouse
	if err := r.db.WithContext(ctx).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("warehouse not found: %w", err)
	}

	return mapToDomainWarehouse(dbModel), nil
}

func (r *WarehouseRepo) List(ctx context.Context) ([]domain.Warehouse, error) {
	var dbModels []models.Warehouse
	if err := r.db.WithContext(ctx).Find(&dbModels).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch warehouses: %w", err)
	}

	var warehouses []domain.Warehouse
	for _, m := range dbModels {
		warehouses = append(warehouses, *mapToDomainWarehouse(m))
	}

	return warehouses, nil
}

func (r *WarehouseRepo) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateWarehouseDTO) error {
	result := r.db.WithContext(ctx).Model(&models.Warehouse{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":      dto.Name,
		"location":  dto.Location,
		"is_active": dto.IsActive,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update warehouse: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("warehouse not found")
	}

	return nil
}

func (r *WarehouseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	// 1. Business logic check: Prevent deletion if there are active lots on this warehouse
	var activeLotsCount int64
	r.db.WithContext(ctx).Model(&models.Lot{}).
		Where("warehouse_id = ? AND current_quantity > 0", id).
		Count(&activeLotsCount)

	if activeLotsCount > 0 {
		return fmt.Errorf("cannot delete warehouse: %d active lots are still stored here", activeLotsCount)
	}

	// 2. Perform Soft Delete (GORM will populate deleted_at column instead of hard delete)
	result := r.db.WithContext(ctx).Delete(&models.Warehouse{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete warehouse: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("warehouse not found")
	}

	return nil
}

func mapToDomainWarehouse(m models.Warehouse) *domain.Warehouse {
	return &domain.Warehouse{
		ID:       m.ID,
		Name:     m.Name,
		Location: m.Location,
		IsActive: m.IsActive,
	}
}
