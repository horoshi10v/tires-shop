package domain

import (
	"context"

	"github.com/google/uuid"
)

// CreateLotDTO contains the necessary data to create a new lot from the API.
type CreateLotDTO struct {
	WarehouseID     uuid.UUID              `json:"warehouse_id" binding:"required"`
	Type            string                 `json:"type" binding:"required,oneof=TIRE RIM"`
	Condition       string                 `json:"condition" binding:"required,oneof=NEW USED"`
	Brand           string                 `json:"brand" binding:"required"`
	Model           string                 `json:"model"`
	Params          map[string]interface{} `json:"params"`
	InitialQuantity int                    `json:"initial_quantity" binding:"required,gt=0"`
	PurchasePrice   float64                `json:"purchase_price" binding:"required,gt=0"`
	SellPrice       float64                `json:"sell_price" binding:"required,gt=0"`
}

// LotFilter defines the criteria for searching and paginating lots.
type LotFilter struct {
	Page     int
	PageSize int
	Status   string
	Brand    string
	Type     string
}

// LotResponse represents the data returned to the client.
// Note: hide PurchasePrice here, as buyers shouldn't see it.
type LotResponse struct {
	ID              uuid.UUID              `json:"id"`
	WarehouseID     uuid.UUID              `json:"warehouse_id"`
	Type            string                 `json:"type"`
	Condition       string                 `json:"condition"`
	Brand           string                 `json:"brand"`
	Model           string                 `json:"model"`
	Params          map[string]interface{} `json:"params"`
	CurrentQuantity int                    `json:"current_quantity"`
	SellPrice       float64                `json:"sell_price"`
	Status          string                 `json:"status"`
}

type LotRepository interface {
	Create(ctx context.Context, dto *CreateLotDTO) (uuid.UUID, error)
	List(ctx context.Context, filter LotFilter) ([]LotResponse, int64, error)
}

type LotService interface {
	CreateLot(ctx context.Context, dto CreateLotDTO) (uuid.UUID, error)
	ListLots(ctx context.Context, filter LotFilter) ([]LotResponse, int64, error)
}
