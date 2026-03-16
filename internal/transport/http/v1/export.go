package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type ExportHandler struct {
	service domain.ExportService
}

func NewExportHandler(service domain.ExportService) *ExportHandler {
	return &ExportHandler{service: service}
}

// ExportInventory creates a Google Sheet with warehouse stock.
//
//	@Summary      Export Inventory to Google Sheets
//	@Tags         exports
//	@Produce      json
//	@Security     RoleAuth
//	@Param        search    query     string  false  "Search by brand or model"
//	@Param        brand     query     string  false  "Filter by brand name"
//	@Param        type      query     string  false  "Filter by type (TIRE, RIM)"
//	@Success      200  {object}  map[string]string "URL of the Google Sheet"
//	@Router       /admin/exports/inventory [get]
func (h *ExportHandler) ExportInventory(c *gin.Context) {
	filter := buildLotFilter(c)

	// Override pagination for export to fetch all (or many) items by default
	if c.Query("page_size") == "" {
		filter.PageSize = 10000
	}

	sheetURL, err := h.service.ExportInventory(c.Request.Context(), filter)
	if err != nil {
		fmt.Println("❌ GOOGLE API ERROR:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"url":     sheetURL,
	})
}

// ExportPnL creates a Google Sheet with Profit and Loss report.
//
//	@Summary      Export P&L to Google Sheets
//	@Tags         exports
//	@Produce      json
//	@Security     RoleAuth
//	@Param        start_date    query     string  false  "Start Date (YYYY-MM-DD)"
//	@Param        end_date      query     string  false  "End Date (YYYY-MM-DD)"
//	@Param        warehouse_id  query     string  false  "Filter by Warehouse ID"
//	@Success      200  {object}  map[string]string "URL of the Google Sheet"
//	@Router       /admin/exports/pnl [get]
func (h *ExportHandler) ExportPnL(c *gin.Context) {
	filter := buildReportFilter(c)

	sheetURL, err := h.service.ExportPnL(c.Request.Context(), filter)
	if err != nil {
		fmt.Println("❌ GOOGLE API ERROR:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"url":     sheetURL,
	})
}
