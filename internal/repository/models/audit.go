package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AuditLog records every significant action in the system for traceability.
type AuditLog struct {
	Base
	Entity   string         `gorm:"type:varchar(50);not null;index"` // e.g., "ORDER", "LOT", "TRANSFER"
	EntityID uuid.UUID      `gorm:"type:uuid;not null;index"`
	UserID   uuid.UUID      `gorm:"type:uuid"`                 // Who did this (Admin/Staff)
	Action   string         `gorm:"type:varchar(50);not null"` // e.g., "STATUS_CHANGED", "PRICE_UPDATED"
	OldValue datatypes.JSON `gorm:"type:jsonb"`                // What it was before
	NewValue datatypes.JSON `gorm:"type:jsonb"`                // What it is now
	Comment  string         `gorm:"type:text"`
}
