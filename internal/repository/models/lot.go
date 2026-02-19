package models

import (
	"github.com/google/uuid"
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

type Lot struct {
	Base
	WarehouseID uuid.UUID `gorm:"type:uuid;not null"` // Where this lot is stored

	// Atributes
	Type      LotType      `gorm:"type:varchar(20);not null"` // TIRE or RIM
	Condition LotCondition `gorm:"type:varchar(20);not null"` // NEW or USED
	Brand     string       `gorm:"type:varchar(100);not null;index"`
	Model     string       `gorm:"type:varchar(100)"`

	// JSON (width: 205, profile: 55, diameter: 16, season: winter)
	Params datatypes.JSON `gorm:"type:jsonb"`

	// Quantitative accounting
	InitialQuantity int `gorm:"not null"` // Came in (4 pcs)
	CurrentQuantity int `gorm:"not null"` // How much is left (2 pcs)

	// Money
	PurchasePrice float64 `gorm:"not null"` // Purchase per piece
	SellPrice     float64 `gorm:"not null"` // Sell per piece

	Status string `gorm:"default:'ACTIVE';index"` // ACTIVE, SOLD, ARCHIVED
}
