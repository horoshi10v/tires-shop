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
//	@Success      200  {object}  map[string]string "URL of the Google Sheet"
//	@Router       /admin/exports/inventory [get]
func (h *ExportHandler) ExportInventory(c *gin.Context) {
	sheetURL, err := h.service.ExportInventory(c.Request.Context())
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
//	@Success      200  {object}  map[string]string "URL of the Google Sheet"
//	@Router       /admin/exports/pnl [get]
func (h *ExportHandler) ExportPnL(c *gin.Context) {
	sheetURL, err := h.service.ExportPnL(c.Request.Context())
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
