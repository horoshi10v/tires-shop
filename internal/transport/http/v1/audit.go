package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type AuditHandler struct {
	service domain.AuditLogService
}

func NewAuditHandler(service domain.AuditLogService) *AuditHandler {
	return &AuditHandler{service: service}
}

// ListAuditLogs retrieves a paginated list of audit logs.
//
//	@Summary      List Audit Logs
//	@Description  Get audit logs with optional filtering by entity, action, user and date range.
//	@Tags         audit-admin
//	@Produce      json
//	@Security     RoleAuth
//	@Param        page       query     int     false  "Page number" default(1)
//	@Param        page_size  query     int     false  "Items per page" default(20)
//	@Param        entity     query     string  false  "Filter by entity"
//	@Param        action     query     string  false  "Filter by action"
//	@Param        user       query     string  false  "Search by user"
//	@Param        start_date query     string  false  "Start date (YYYY-MM-DD)"
//	@Param        end_date   query     string  false  "End date (YYYY-MM-DD)"
//	@Success      200        {array}   domain.AuditLogResponse
//	@Router       /admin/audit-logs [get]
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var startDate *time.Time
	if rawStartDate := c.Query("start_date"); rawStartDate != "" {
		parsedStartDate, err := time.Parse("2006-01-02", rawStartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date, expected YYYY-MM-DD"})
			return
		}
		startDate = &parsedStartDate
	}

	var endDate *time.Time
	if rawEndDate := c.Query("end_date"); rawEndDate != "" {
		parsedEndDate, err := time.Parse("2006-01-02", rawEndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date, expected YYYY-MM-DD"})
			return
		}
		endDate = &parsedEndDate
	}

	filter := domain.AuditLogFilter{
		Page:      page,
		PageSize:  pageSize,
		Entity:    c.Query("entity"),
		Action:    c.Query("action"),
		User:      c.Query("user"),
		StartDate: startDate,
		EndDate:   endDate,
	}

	logs, total, err := h.service.ListAuditLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit logs"})
		return
	}

	if logs == nil {
		logs = []domain.AuditLogResponse{}
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count")
	c.JSON(http.StatusOK, logs)
}
