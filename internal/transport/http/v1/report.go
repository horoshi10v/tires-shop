package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
//	@Success      200  {object}  domain.PnLReport
//	@Failure      401  {object}  map[string]string "Unauthorized"
//	@Failure      403  {object}  map[string]string "Forbidden"
//	@Router       /reports/pnl [get]
func (h *ReportHandler) GetPnL(c *gin.Context) {
	report, err := h.service.GetPnLReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, report)
}
