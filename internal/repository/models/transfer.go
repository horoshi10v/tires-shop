package models

import (
	"github.com/google/uuid"
)

// Transfer represents a movement of stock between two warehouses.
type Transfer struct {
	Base
	FromWarehouseID uuid.UUID  `gorm:"type:uuid;not null;index"`
	ToWarehouseID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	Status          string     `gorm:"type:varchar(20);default:'IN_TRANSIT';index"`
	CreatedByID     uuid.UUID  `gorm:"type:uuid;not null"`
	AcceptedByID    *uuid.UUID `gorm:"type:uuid"` // Nullable, filled when accepted
	Comment         string     `gorm:"type:text"`

	// Has-Many relationship
	Items []TransferItem `gorm:"foreignKey:TransferID"`
}

// TransferItem represents a specific portion of a lot being moved.
type TransferItem struct {
	Base
	TransferID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	SourceLotID      uuid.UUID  `gorm:"type:uuid;not null"`
	DestinationLotID *uuid.UUID `gorm:"type:uuid"` // Filled when transfer is ACCEPTED (the newly created lot)
	Quantity         int        `gorm:"not null"`
}
