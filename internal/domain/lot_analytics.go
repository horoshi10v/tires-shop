package domain

import "github.com/google/uuid"

type LotAnalyticsEventType string

type LotAnalyticsSource string

const (
	LotAnalyticsEventView           LotAnalyticsEventType = "VIEW"
	LotAnalyticsEventFavoriteAdd    LotAnalyticsEventType = "FAVORITE_ADD"
	LotAnalyticsEventFavoriteRemove LotAnalyticsEventType = "FAVORITE_REMOVE"
	LotAnalyticsEventOrderCreated   LotAnalyticsEventType = "ORDER_CREATED"
)

const (
	LotAnalyticsSourceWeb   LotAnalyticsSource = "WEB"
	LotAnalyticsSourceTMA   LotAnalyticsSource = "TMA"
	LotAnalyticsSourceStaff LotAnalyticsSource = "STAFF"
)

type TrackLotAnalyticsEventRequest struct {
	LotID     uuid.UUID             `json:"lot_id" binding:"required"`
	EventType LotAnalyticsEventType `json:"event_type" binding:"required,oneof=VIEW FAVORITE_ADD FAVORITE_REMOVE"`
	Source    LotAnalyticsSource    `json:"source" binding:"required,oneof=WEB TMA STAFF"`
	SessionID string                `json:"session_id" binding:"omitempty,max=120"`
}
