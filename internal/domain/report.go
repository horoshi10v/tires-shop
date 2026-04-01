package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PnLReport represents the Profit and Loss financial data.
type PnLReport struct {
	TotalItemsSold int            `json:"total_items_sold"`
	TotalRevenue   float64        `json:"total_revenue"`
	TotalCOGS      float64        `json:"total_cogs"`
	TotalProfit    float64        `json:"total_profit"`
	ByWarehouse    []WarehousePnL `json:"by_warehouse"`
	ByChannel      []ChannelPnL   `json:"by_channel"`
}

type LotAnalyticsTotals struct {
	Views          int     `json:"views"`
	FavoritesAdded int     `json:"favorites_added"`
	OrdersCreated  int     `json:"orders_created"`
	ConversionRate float64 `json:"conversion_rate"`
}

type LotAnalyticsDailyPoint struct {
	Date           string `json:"date"`
	Views          int    `json:"views"`
	FavoritesAdded int    `json:"favorites_added"`
	OrdersCreated  int    `json:"orders_created"`
}

type LotAnalyticsLotRow struct {
	LotID          uuid.UUID `json:"lot_id"`
	Brand          string    `json:"brand"`
	Model          string    `json:"model"`
	Type           string    `json:"type"`
	Condition      string    `json:"condition"`
	Views          int       `json:"views"`
	FavoritesAdded int       `json:"favorites_added"`
	OrdersCreated  int       `json:"orders_created"`
	ConversionRate float64   `json:"conversion_rate"`
}

type LotAnalyticsReport struct {
	Totals        LotAnalyticsTotals       `json:"totals"`
	Daily         []LotAnalyticsDailyPoint `json:"daily"`
	TopViewed     []LotAnalyticsLotRow     `json:"top_viewed"`
	TopFavorited  []LotAnalyticsLotRow     `json:"top_favorited"`
	TopConverting []LotAnalyticsLotRow     `json:"top_converting"`
}

// ReportFilter defines criteria for filtering reports.
type ReportFilter struct {
	StartDate   *time.Time
	EndDate     *time.Time
	WarehouseID *uuid.UUID
	Channel     *OrderChannel
	LotID       *uuid.UUID
	Type        *string
	Source      *LotAnalyticsSource
}

// ReportRepository handles analytical database queries.
type ReportRepository interface {
	GetPnL(ctx context.Context, filter ReportFilter) (*PnLReport, error)
	GetLotAnalytics(ctx context.Context, filter ReportFilter) (*LotAnalyticsReport, error)
}

// ReportService handles business logic for reports.
type ReportService interface {
	GetPnLReport(ctx context.Context, filter ReportFilter) (*PnLReport, error)
	GetLotAnalyticsReport(ctx context.Context, filter ReportFilter) (*LotAnalyticsReport, error)
}

// WarehousePnL contains finances per warehouse.
type WarehousePnL struct {
	WarehouseName string  `json:"warehouse_name"`
	ItemsSold     int     `json:"items_sold"`
	Revenue       float64 `json:"revenue"`
	COGS          float64 `json:"cogs"`
	Profit        float64 `json:"profit"`
}

// ChannelPnL contains financial metrics grouped by sales channel.
type ChannelPnL struct {
	Channel   OrderChannel `json:"channel"`
	ItemsSold int          `json:"items_sold"`
	Revenue   float64      `json:"revenue"`
	COGS      float64      `json:"cogs"`
	Profit    float64      `json:"profit"`
}
