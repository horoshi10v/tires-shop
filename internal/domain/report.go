package domain

import "context"

// PnLReport represents the Profit and Loss financial data.
type PnLReport struct {
	Revenue float64 `json:"revenue"` // Total sales amount
	COGS    float64 `json:"cogs"`    // Cost Of Goods Sold (Purchase price)
	Profit  float64 `json:"profit"`  // Revenue - COGS
}

// ReportRepository handles analytical database queries.
type ReportRepository interface {
	GetPnL(ctx context.Context) (*PnLReport, error)
}

// ReportService handles business logic for reports.
type ReportService interface {
	GetPnLReport(ctx context.Context) (*PnLReport, error)
}
