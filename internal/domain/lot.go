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
type LotParams struct {
	Width             int    `json:"width,omitempty"`
	Profile           int    `json:"profile,omitempty"`
	Diameter          int    `json:"diameter,omitempty"`
	ProductionYear    int    `json:"production_year,omitempty"`
	CountryOfOrigin   string `json:"country_of_origin,omitempty"`
	Season            string `json:"season,omitempty"` // SUMMER, WINTER, ALL_SEASON
	IsRunFlat         bool   `json:"is_run_flat"`
	IsSpiked          bool   `json:"is_spiked"`
	AntiPuncture      bool   `json:"anti_puncture"`
	AccessoryCategory string `json:"accessory_category,omitempty"` // FASTENERS, HUB_RINGS, SPACERS, TIRE_BAGS
	FastenerType      string `json:"fastener_type,omitempty"`      // NUT, BOLT
	ThreadSize        string `json:"thread_size,omitempty"`
	SeatType          string `json:"seat_type,omitempty"`
	RingInnerDiameter int    `json:"ring_inner_diameter,omitempty"`
	RingOuterDiameter int    `json:"ring_outer_diameter,omitempty"`
	SpacerType        string `json:"spacer_type,omitempty"` // ADAPTER, EXTENDER
	SpacerThickness   int    `json:"spacer_thickness,omitempty"`
	PackageQuantity   int    `json:"package_quantity,omitempty"`
}

// CreateLotDTO contains the necessary data to create a new lot from the API.
type CreateLotDTO struct {
	WarehouseID     uuid.UUID `json:"warehouse_id" binding:"required"`
	Type            string    `json:"type" binding:"required,oneof=TIRE RIM ACCESSORY"`
	Condition       string    `json:"condition" binding:"required,oneof=NEW USED"`
	Brand           string    `json:"brand" binding:"required"`
	Model           string    `json:"model"`
	Params          LotParams `json:"params"`
	Defects         string    `json:"defects"`
	Photos          []string  `json:"photos"`
	InitialQuantity int       `json:"initial_quantity" binding:"required,gt=0"`
	PurchasePrice   float64   `json:"purchase_price" binding:"required,gt=0"`
	SellPrice       float64   `json:"sell_price" binding:"required,gt=0"`
}

// UpdateLotDTO contains fields that can be updated.
type UpdateLotDTO struct {
	WarehouseID   *uuid.UUID `json:"warehouse_id"`
	Type          *string    `json:"type" binding:"omitempty,oneof=TIRE RIM ACCESSORY"`
	Condition     *string    `json:"condition" binding:"omitempty,oneof=NEW USED"`
	Brand         *string    `json:"brand"`
	Model         *string    `json:"model"`
	Params        *LotParams `json:"params"`
	Defects       *string    `json:"defects"`
	Photos        []string   `json:"photos"`
	PurchasePrice *float64   `json:"purchase_price" binding:"omitempty,gt=0"`
	SellPrice     *float64   `json:"sell_price" binding:"omitempty,gt=0"`
}

// LotFilter defines the criteria for searching and paginating lots.
type LotFilter struct {
	Page            int
	PageSize        int
	SortBy          string
	SortOrder       string
	Status          string
	Brand           string
	Type            string
	Search          string
	Width           int
	Profile         int
	Diameter        int
	ProductionYear  int
	CountryOfOrigin string
	Season          string

	IsRunFlat    *bool
	IsSpiked     *bool
	AntiPuncture *bool

	Condition         string
	Model             string
	CurrentQuantity   *int
	SellPrice         *float64
	AccessoryCategory string
	FastenerType      string
	ThreadSize        string
	SeatType          string
	RingInnerDiameter int
	RingOuterDiameter int
	SpacerType        string
	SpacerThickness   int
	PackageQuantity   int
}

// LotPublicResponse is what the BUYER sees.
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

// PaginatedLotPublicResponse is the paginated public contract for /lots.
type PaginatedLotPublicResponse struct {
	Items    []LotPublicResponse `json:"items"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Total    int64               `json:"total"`
	HasNext  bool                `json:"has_next"`
}

// LotInternalResponse is what ADMIN and STAFF see.
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
	Update(ctx context.Context, id uuid.UUID, dto *UpdateLotDTO) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListPublic(ctx context.Context, filter LotFilter) ([]LotPublicResponse, int64, error)
	ListInternal(ctx context.Context, filter LotFilter) ([]LotInternalResponse, int64, error)
}

// LotService defines business logic operations for the Lot entity.
type LotService interface {
	CreateLot(ctx context.Context, dto CreateLotDTO) (uuid.UUID, error)
	UpdateLot(ctx context.Context, id uuid.UUID, dto UpdateLotDTO) error
	DeleteLot(ctx context.Context, id uuid.UUID) error
	ListPublicLots(ctx context.Context, filter LotFilter) ([]LotPublicResponse, int64, error)
	ListInternalLots(ctx context.Context, filter LotFilter) ([]LotInternalResponse, int64, error)
	GenerateLotQR(ctx context.Context, id uuid.UUID) ([]byte, error)
}
