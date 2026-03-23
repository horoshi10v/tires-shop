package models

import "github.com/google/uuid"

type OrderMessage struct {
	Base
	OrderID                  uuid.UUID `gorm:"type:uuid;not null;index"`
	CustomerTelegramID       int64     `gorm:"not null;index"`
	Direction                string    `gorm:"type:varchar(20);not null;index"`
	MessageText              string    `gorm:"type:text;not null"`
	TelegramMessageID        int64     `gorm:"not null;index"`
	ReplyToTelegramMessageID *int64    `gorm:"index"`
}
