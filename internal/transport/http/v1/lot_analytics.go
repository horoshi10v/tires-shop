package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

func (h *LotHandler) TrackPublicAnalyticsEvent(c *gin.Context) {
	var req domain.TrackLotAnalyticsEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	if err := h.service.TrackLotAnalyticsEvent(c.Request.Context(), req, c.GetHeader("User-Agent")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to track lot analytics event"})
		return
	}

	c.Status(http.StatusNoContent)
}
