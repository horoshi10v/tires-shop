package domain

import (
	"context"

	"github.com/google/uuid"
)

// Warehouse represents the business entity of a storage location.
type Warehouse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Location string    `json:"location"`
	IsActive bool      `json:"is_active"`
}

// CreateWarehouseDTO represents the payload for creating a warehouse.
type CreateWarehouseDTO struct {
	Name     string `json:"name" binding:"required,min=3"`
	Location string `json:"location" binding:"required"`
}

// UpdateWarehouseDTO represents the payload for updating a warehouse.
type UpdateWarehouseDTO struct {
	Name     string `json:"name" binding:"required,min=3"`
	Location string `json:"location" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// WarehouseRepository defines data access methods for warehouses.
type WarehouseRepository interface {
	Create(ctx context.Context, dto CreateWarehouseDTO) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Warehouse, error)
	List(ctx context.Context) ([]Warehouse, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateWarehouseDTO) error
	Delete(ctx context.Context, id uuid.UUID) error // Soft delete
}

// WarehouseService defines business logic for warehouses.
type WarehouseService interface {
	CreateWarehouse(ctx context.Context, dto CreateWarehouseDTO) (uuid.UUID, error)
	GetWarehouse(ctx context.Context, id uuid.UUID) (*Warehouse, error)
	ListWarehouses(ctx context.Context) ([]Warehouse, error)
	UpdateWarehouse(ctx context.Context, id uuid.UUID, dto UpdateWarehouseDTO) error
	DeleteWarehouse(ctx context.Context, id uuid.UUID) error
}
