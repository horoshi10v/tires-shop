package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type ReportHandler struct {
	service domain.ReportService
}

func NewReportHandler(service domain.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

// GetPnL returns the Profit and Loss financial report.
//
//	@Summary      Get Profit & Loss Report
//	@Description  Calculates total revenue, COGS, and profit based on DONE orders.
//	@Tags         reports
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        start_date    query     string  false  "Start Date (YYYY-MM-DD)"
//	@Param        end_date      query     string  false  "End Date (YYYY-MM-DD)"
//	@Param        warehouse_id  query     string  false  "Filter by Warehouse ID"
//	@Param        channel       query     string  false  "Filter by sales channel (ONLINE|OFFLINE)"
//	@Success      200  {object}  domain.PnLReport
//	@Failure      401  {object}  map[string]string "Unauthorized"
//	@Failure      403  {object}  map[string]string "Forbidden"
//	@Router       /reports/pnl [get]
func (h *ReportHandler) GetPnL(c *gin.Context) {
	filter := buildReportFilter(c)

	report, err := h.service.GetPnLReport(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

func buildReportFilter(c *gin.Context) domain.ReportFilter {
	var startDate *time.Time
	if val := c.Query("start_date"); val != "" {
		if t, err := time.Parse("2006-01-02", val); err == nil {
			startDate = &t
		}
	}

	var endDate *time.Time
	if val := c.Query("end_date"); val != "" {
		if t, err := time.Parse("2006-01-02", val); err == nil {
			// Set time to end of day 23:59:59
			t = t.Add(24*time.Hour - time.Nanosecond)
			endDate = &t
		}
	}

	var warehouseID *uuid.UUID
	if val := c.Query("warehouse_id"); val != "" {
		if id, err := uuid.Parse(val); err == nil {
			warehouseID = &id
		}
	}

	var channel *domain.OrderChannel
	if val := c.Query("channel"); val != "" {
		parsed := domain.OrderChannel(val)
		channel = &parsed
	}

	return domain.ReportFilter{
		StartDate:   startDate,
		EndDate:     endDate,
		WarehouseID: warehouseID,
		Channel:     channel,
	}
}
