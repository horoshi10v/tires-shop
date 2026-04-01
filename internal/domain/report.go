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

// ReportFilter defines criteria for filtering reports.
type ReportFilter struct {
	StartDate   *time.Time
	EndDate     *time.Time
	WarehouseID *uuid.UUID
	Channel     *OrderChannel
}

// ReportRepository handles analytical database queries.
type ReportRepository interface {
	GetPnL(ctx context.Context, filter ReportFilter) (*PnLReport, error)
}

// ReportService handles business logic for reports.
type ReportService interface {
	GetPnLReport(ctx context.Context, filter ReportFilter) (*PnLReport, error)
}

// WarehousePnL содержит финансы по одному конкретному складу
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
