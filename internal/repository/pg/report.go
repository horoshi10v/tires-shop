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
func (r *ReportRepo) GetPnL(ctx context.Context, filter domain.ReportFilter) (*domain.PnLReport, error) {
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
	`

	var args []interface{}

	if filter.StartDate != nil {
		query += " AND o.created_at >= ?"
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		query += " AND o.created_at <= ?"
		args = append(args, *filter.EndDate)
	}

	if filter.WarehouseID != nil {
		query += " AND w.id = ?"
		args = append(args, *filter.WarehouseID)
	}

	if filter.Channel != nil {
		query += " AND o.channel = ?"
		args = append(args, string(*filter.Channel))
	}

	query += `
		GROUP BY w.id, w.name
		ORDER BY w.name
	`

	var warehousePnLs []domain.WarehousePnL
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&warehousePnLs).Error; err != nil {
		return nil, err
	}

	channelQuery := `
		SELECT
			o.channel as channel,
			COALESCE(SUM(oi.quantity), 0) as items_sold,
			COALESCE(SUM(oi.quantity * oi.price_at_moment), 0) as revenue,
			COALESCE(SUM(oi.quantity * oi.cost_at_moment), 0) as cogs,
			COALESCE(SUM(oi.quantity * (oi.price_at_moment - oi.cost_at_moment)), 0) as profit
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN lots l ON oi.lot_id = l.id
		JOIN warehouses w ON l.warehouse_id = w.id
		WHERE o.status = 'DONE' AND o.deleted_at IS NULL
	`

	channelArgs := append([]interface{}{}, args...)
	if filter.Channel == nil {
		// nothing
	}
	if filter.StartDate != nil {
		// already encoded in args ordering above
	}
	if filter.EndDate != nil {
		// already encoded in args ordering above
	}
	if filter.WarehouseID != nil {
		// already encoded in args ordering above
	}

	if filter.StartDate != nil {
		channelQuery += " AND o.created_at >= ?"
	}
	if filter.EndDate != nil {
		channelQuery += " AND o.created_at <= ?"
	}
	if filter.WarehouseID != nil {
		channelQuery += " AND w.id = ?"
	}
	if filter.Channel != nil {
		channelQuery += " AND o.channel = ?"
	}

	channelQuery += `
		GROUP BY o.channel
		ORDER BY o.channel
	`

	var channelPnLs []domain.ChannelPnL
	if err := r.db.WithContext(ctx).Raw(channelQuery, channelArgs...).Scan(&channelPnLs).Error; err != nil {
		return nil, err
	}

	report := &domain.PnLReport{
		ByWarehouse:    warehousePnLs,
		ByChannel:      channelPnLs,
		TotalItemsSold: 0,
		TotalRevenue:   0,
		TotalCOGS:      0,
		TotalProfit:    0,
	}

	for _, wpnl := range warehousePnLs {
		report.TotalItemsSold += wpnl.ItemsSold
		report.TotalRevenue += wpnl.Revenue
		report.TotalCOGS += wpnl.COGS
		report.TotalProfit += wpnl.Profit
	}

	return report, nil
}
