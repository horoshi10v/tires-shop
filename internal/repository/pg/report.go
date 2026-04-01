package pg

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/horoshi10v/tires-shop/internal/domain"
)

type ReportRepo struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) domain.ReportRepository {
	return &ReportRepo{db: db}
}

func buildAnalyticsConditions(filter domain.ReportFilter) (string, []interface{}) {
	conditions := []string{"lae.deleted_at IS NULL", "l.deleted_at IS NULL"}
	args := make([]interface{}, 0, 4)

	if filter.StartDate != nil {
		conditions = append(conditions, "lae.created_at >= ?")
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		conditions = append(conditions, "lae.created_at <= ?")
		args = append(args, *filter.EndDate)
	}
	if filter.WarehouseID != nil {
		conditions = append(conditions, "l.warehouse_id = ?")
		args = append(args, *filter.WarehouseID)
	}
	if filter.LotID != nil {
		conditions = append(conditions, "lae.lot_id = ?")
		args = append(args, *filter.LotID)
	}
	if filter.Type != nil && *filter.Type != "" {
		conditions = append(conditions, "l.type = ?")
		args = append(args, *filter.Type)
	}
	if filter.Source != nil && *filter.Source != "" {
		conditions = append(conditions, "lae.source = ?")
		args = append(args, string(*filter.Source))
	}

	return strings.Join(conditions, " AND "), args
}

// GetPnL executes a raw SQL query to calculate financials.
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

func (r *ReportRepo) GetLotAnalytics(ctx context.Context, filter domain.ReportFilter) (*domain.LotAnalyticsReport, error) {
	analyticsConditions, analyticsArgs := buildAnalyticsConditions(filter)

	var totals domain.LotAnalyticsTotals
	totalsQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN lae.event_type = 'VIEW' THEN 1 ELSE 0 END), 0) AS views,
			COALESCE(SUM(CASE WHEN lae.event_type = 'FAVORITE_ADD' THEN 1 ELSE 0 END), 0) AS favorites_added,
			COALESCE(SUM(CASE WHEN lae.event_type = 'ORDER_CREATED' THEN 1 ELSE 0 END), 0) AS orders_created
		FROM lot_analytics_events lae
		JOIN lots l ON l.id = lae.lot_id
		WHERE ` + analyticsConditions

	if err := r.db.WithContext(ctx).Raw(totalsQuery, analyticsArgs...).Scan(&totals).Error; err != nil {
		return nil, err
	}
	if totals.Views > 0 {
		totals.ConversionRate = float64(totals.OrdersCreated) / float64(totals.Views)
	}

	var daily []domain.LotAnalyticsDailyPoint
	dailyQuery := `
		SELECT
			TO_CHAR(DATE(lae.created_at), 'YYYY-MM-DD') AS date,
			COALESCE(SUM(CASE WHEN lae.event_type = 'VIEW' THEN 1 ELSE 0 END), 0) AS views,
			COALESCE(SUM(CASE WHEN lae.event_type = 'FAVORITE_ADD' THEN 1 ELSE 0 END), 0) AS favorites_added,
			COALESCE(SUM(CASE WHEN lae.event_type = 'ORDER_CREATED' THEN 1 ELSE 0 END), 0) AS orders_created
		FROM lot_analytics_events lae
		JOIN lots l ON l.id = lae.lot_id
		WHERE ` + analyticsConditions + `
		GROUP BY DATE(lae.created_at)
		ORDER BY DATE(lae.created_at) ASC
	`
	if err := r.db.WithContext(ctx).Raw(dailyQuery, analyticsArgs...).Scan(&daily).Error; err != nil {
		return nil, err
	}

	rowSelect := `
		SELECT
			lae.lot_id,
			l.brand,
			l.model,
			l.type,
			l.condition,
			COALESCE(SUM(CASE WHEN lae.event_type = 'VIEW' THEN 1 ELSE 0 END), 0) AS views,
			COALESCE(SUM(CASE WHEN lae.event_type = 'FAVORITE_ADD' THEN 1 ELSE 0 END), 0) AS favorites_added,
			COALESCE(SUM(CASE WHEN lae.event_type = 'ORDER_CREATED' THEN 1 ELSE 0 END), 0) AS orders_created,
			CASE
				WHEN COALESCE(SUM(CASE WHEN lae.event_type = 'VIEW' THEN 1 ELSE 0 END), 0) > 0 THEN
					COALESCE(SUM(CASE WHEN lae.event_type = 'ORDER_CREATED' THEN 1 ELSE 0 END), 0)::float /
					COALESCE(SUM(CASE WHEN lae.event_type = 'VIEW' THEN 1 ELSE 0 END), 0)::float
				ELSE 0
			END AS conversion_rate
		FROM lot_analytics_events lae
		JOIN lots l ON l.id = lae.lot_id
		WHERE ` + analyticsConditions + `
		GROUP BY lae.lot_id, l.brand, l.model, l.type, l.condition
	`

	loadRows := func(orderBy string) ([]domain.LotAnalyticsLotRow, error) {
		var rows []domain.LotAnalyticsLotRow
		query := rowSelect + ` ORDER BY ` + orderBy + ` LIMIT 10`
		if err := r.db.WithContext(ctx).Raw(query, analyticsArgs...).Scan(&rows).Error; err != nil {
			return nil, err
		}
		return rows, nil
	}

	topViewed, err := loadRows("views DESC, orders_created DESC, favorites_added DESC, l.brand ASC, l.model ASC")
	if err != nil {
		return nil, err
	}
	topFavorited, err := loadRows("favorites_added DESC, views DESC, orders_created DESC, l.brand ASC, l.model ASC")
	if err != nil {
		return nil, err
	}
	topConverting, err := loadRows("conversion_rate DESC, orders_created DESC, views DESC, l.brand ASC, l.model ASC")
	if err != nil {
		return nil, err
	}

	return &domain.LotAnalyticsReport{
		Totals:        totals,
		Daily:         daily,
		TopViewed:     topViewed,
		TopFavorited:  topFavorited,
		TopConverting: topConverting,
	}, nil
}
