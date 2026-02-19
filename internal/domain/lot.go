package domain

import (
	"context"

	"github.com/google/uuid"
)

// LotStatus defines the lifecycle of a lot.
type LotStatus string

const (
	LotStatusActive   LotStatus = "ACTIVE"
	LotStatusReserved LotStatus = "RESERVED"
	LotStatusArchived LotStatus = "ARCHIVED" // After full sale
)

// LotParams contains specific attributes for tires/rims.
// Using a struct instead of map[string]interface{} provides strict type checking.
type LotParams struct {
	Width        int    `json:"width,omitempty"`
	Profile      int    `json:"profile,omitempty"`
	Diameter     int    `json:"diameter,omitempty"`
	Season       string `json:"season,omitempty"` // SUMMER, WINTER, ALL_SEASON
	IsRunFlat    bool   `json:"is_run_flat"`
	IsSpiked     bool   `json:"is_spiked"`
	AntiPuncture bool   `json:"anti_puncture"`
}

// CreateLotDTO contains the necessary data to create a new lot from the API.
type CreateLotDTO struct {
	WarehouseID     uuid.UUID `json:"warehouse_id" binding:"required"`
	Type            string    `json:"type" binding:"required,oneof=TIRE RIM"`
	Condition       string    `json:"condition" binding:"required,oneof=NEW USED"`
	Brand           string    `json:"brand" binding:"required"`
	Model           string    `json:"model"`
	Params          LotParams `json:"params"`
	Defects         string    `json:"defects"` // Description of damages, if any
	Photos          []string  `json:"photos"`  // URLs to images
	InitialQuantity int       `json:"initial_quantity" binding:"required,gt=0"`
	PurchasePrice   float64   `json:"purchase_price" binding:"required,gt=0"` // Hidden from buyer
	SellPrice       float64   `json:"sell_price" binding:"required,gt=0"`
}

// LotFilter defines the criteria for searching and paginating lots.
type LotFilter struct {
	Page     int
	PageSize int
	Status   string
	Brand    string
	Type     string
}

// LotPublicResponse is what the BUYER sees.
// Note: PurchasePrice, InitialQuantity, and Warehouse details are hidden.
type LotPublicResponse struct {
	ID              uuid.UUID `json:"id"`
	Type            string    `json:"type"`
	Condition       string    `json:"condition"`
	Brand           string    `json:"brand"`
	Model           string    `json:"model"`
	Params          LotParams `json:"params"`
	Defects         string    `json:"defects,omitempty"`
	Photos          []string  `json:"photos"`
	CurrentQuantity int       `json:"current_quantity"`
	SellPrice       float64   `json:"sell_price"`
}

// LotInternalResponse is what ADMIN and STAFF see.
// It embeds the public response and adds sensitive financial/warehouse data.
type LotInternalResponse struct {
	LotPublicResponse
	WarehouseID   uuid.UUID `json:"warehouse_id"`
	InitialQty    int       `json:"initial_quantity"`
	PurchasePrice float64   `json:"purchase_price"`
	Status        string    `json:"status"`
}

// LotRepository defines database operations for the Lot entity.
type LotRepository interface {
	Create(ctx context.Context, dto *CreateLotDTO) (uuid.UUID, error)
	ListPublic(ctx context.Context, filter LotFilter) ([]LotPublicResponse, int64, error)
	ListInternal(ctx context.Context, filter LotFilter) ([]LotInternalResponse, int64, error)
}

// LotService defines business logic operations for the Lot entity.
type LotService interface {
	CreateLot(ctx context.Context, dto CreateLotDTO) (uuid.UUID, error)
	ListPublicLots(ctx context.Context, filter LotFilter) ([]LotPublicResponse, int64, error)
	ListInternalLots(ctx context.Context, filter LotFilter) ([]LotInternalResponse, int64, error)
}
