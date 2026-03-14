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
	query := `
		SELECT 
			w.name as warehouse_name,
			COALESCE(SUM(oi.quantity), 0) as items_sold,
			COALESCE(SUM(oi.quantity * oi.price_at_moment), 0) as revenue,
			COALESCE(SUM(oi.quantity * oi.cost_at_moment), 0) as cogs,
			COALESCE(SUM(oi.quantity * (oi.price_at_moment - oi.cost_at_moment)), 0) as profit
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN lots l ON oi.lot_id = l.id
		JOIN warehouses w ON l.warehouse_id = w.id
		WHERE o.status = 'DONE' AND o.deleted_at IS NULL
		GROUP BY w.id, w.name
		ORDER BY w.name
	`

	var warehousePnLs []domain.WarehousePnL
	if err := r.db.WithContext(ctx).Raw(query).Scan(&warehousePnLs).Error; err != nil {
		return nil, err
	}

	report := &domain.PnLReport{
		ByWarehouse: warehousePnLs,
	}

	for _, wpnl := range warehousePnLs {
		report.TotalItemsSold += wpnl.ItemsSold
		report.TotalRevenue += wpnl.Revenue
		report.TotalCOGS += wpnl.COGS
		report.TotalProfit += wpnl.Profit
	}

	return report, nil
}
