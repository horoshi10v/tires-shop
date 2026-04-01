package models

import "github.com/google/uuid"

type LotAnalyticsEvent struct {
	Base
	LotID     uuid.UUID `gorm:"type:uuid;not null;index:idx_lot_analytics_lot_created"`
	EventType string    `gorm:"type:varchar(32);not null;index:idx_lot_analytics_type_created"`
	Source    string    `gorm:"type:varchar(16);not null;index"`
	SessionID string    `gorm:"type:varchar(120);index"`
	UserAgent string    `gorm:"type:text"`
}
