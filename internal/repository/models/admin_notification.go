package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AdminNotification struct {
	Base
	Type               string         `gorm:"type:varchar(40);not null;index"`
	Title              string         `gorm:"type:varchar(255);not null"`
	Body               string         `gorm:"type:text;not null"`
	OrderID            *uuid.UUID     `gorm:"type:uuid;index"`
	CustomerName       string         `gorm:"type:varchar(120)"`
	CustomerPhone      string         `gorm:"type:varchar(30)"`
	CustomerUsername   string         `gorm:"type:varchar(120)"`
	CustomerTelegramID *int64         `gorm:"index"`
	Payload            datatypes.JSON `gorm:"type:jsonb"`
	IsRead             bool           `gorm:"not null;default:false;index"`
}
