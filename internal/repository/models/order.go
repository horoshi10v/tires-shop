package models

import (
	"github.com/google/uuid"
)

// Order represents the main order document in the database.
type Order struct {
	Base
	UserID             *uuid.UUID `gorm:"type:uuid;index"` // Nullable for guest orders, but currently all are auth'd
	CustomerName       string     `gorm:"type:varchar(100)"`
	CustomerPhone      string     `gorm:"type:varchar(20);index"`
	CustomerUsername   string     `gorm:"type:varchar(100);index"`              // Added for Telegram integration
	CustomerTelegramID *int64     `gorm:"index"`                                // Added for Telegram integration
	Status             string     `gorm:"type:varchar(20);default:'NEW';index"` // NEW, PREPAYMENT, DONE, CANCELLED
	TotalAmount        float64    `gorm:"not null"`

	// Has-Many relationship
	Items []OrderItem `gorm:"foreignKey:OrderID"`
}

// OrderItem represents a specific lot deducted for an order.
type OrderItem struct {
	Base
	OrderID uuid.UUID `gorm:"type:uuid;not null;index"`
	LotID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Brand   string    `gorm:"type:varchar(100)"`
	Model   string    `gorm:"type:varchar(100)"`
	Photo   string    `gorm:"type:text"`

	Quantity      int     `gorm:"not null"`
	PriceAtMoment float64 `gorm:"not null"` // Sell price at the time of order
	CostAtMoment  float64 `gorm:"not null"` // Purchase price at the time of order (for P&L)
}
