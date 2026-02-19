package pg

import (
	"context"

	"gorm.io/gorm"

	"github.com/horoshi10v/tires-shop/internal/domain"
)

type ReportRepo struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) domain.ReportRepository {
	return &ReportRepo{db: db}
}

// GetPnL executes a raw SQL query to calculate financials.
// This demonstrates ability to write complex analytical queries manually.
func (r *ReportRepo) GetPnL(ctx context.Context) (*domain.PnLReport, error) {
	var report domain.PnLReport

	// The SQL query calculates Revenue, COGS, and Profit across all DONE orders.
	// COALESCE is used to return 0 instead of NULL if there are no finished orders yet.
	query := `
		SELECT 
			COALESCE(SUM(oi.quantity * oi.price_at_moment), 0) as revenue,
			COALESCE(SUM(oi.quantity * oi.cost_at_moment), 0) as cogs,
			COALESCE(SUM(oi.quantity * (oi.price_at_moment - oi.cost_at_moment)), 0) as profit
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.status = 'DONE' AND o.deleted_at IS NULL
	`

	if err := r.db.WithContext(ctx).Raw(query).Scan(&report).Error; err != nil {
		return nil, err
	}

	return &report, nil
}
