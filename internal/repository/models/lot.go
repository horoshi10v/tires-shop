package models

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type LotType string
type LotCondition string

const (
	LotTypeTire LotType = "TIRE"
	LotTypeRim  LotType = "RIM"

	ConditionNew  LotCondition = "NEW"
	ConditionUsed LotCondition = "USED"
)

// Lot represents the database schema for the lots table.
type Lot struct {
	Base
	WarehouseID uuid.UUID `gorm:"type:uuid;not null;index"` // Where this lot is stored

	// Attributes
	Type      LotType      `gorm:"type:varchar(20);not null"` // TIRE or RIM
	Condition LotCondition `gorm:"type:varchar(20);not null"` // NEW or USED
	Brand     string       `gorm:"type:varchar(100);not null;index"`
	Model     string       `gorm:"type:varchar(100)"`

	// JSONB strongly typed to domain.LotParams in the application layer
	Params datatypes.JSON `gorm:"type:jsonb"`

	// Additional details required by the PRD
	Defects string         `gorm:"type:text"`
	Photos  pq.StringArray `gorm:"type:text[]"` // Native PostgreSQL array for photo URLs

	// Quantitative accounting (Lot Model concept)
	InitialQuantity int `gorm:"not null"`       // Came in (e.g., 4 pcs)
	CurrentQuantity int `gorm:"not null;index"` // How much is left (e.g., 2 pcs) - Indexed for fast queries

	// Money
	PurchasePrice float64 `gorm:"not null"` // Purchase per piece (hidden from buyer)
	SellPrice     float64 `gorm:"not null"` // Sell per piece

	Status string `gorm:"type:varchar(20);default:'ACTIVE';index"` // ACTIVE, SOLD, ARCHIVED
}
