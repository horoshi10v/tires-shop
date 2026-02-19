package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type LotHandler struct {
	service domain.LotService
}

func NewLotHandler(service domain.LotService) *LotHandler {
	return &LotHandler{service: service}
}

func (h *LotHandler) Create(c *gin.Context) {
	var req domain.CreateLotDTO

	// Bind JSON body to struct and validate tags (e.g. required, gt=0)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Pass context and validated DTO to the service layer
	id, err := h.service.CreateLot(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Return 201 Created on success
	c.JSON(http.StatusCreated, gin.H{
		"message": "lot created successfully",
		"lot_id":  id,
	})
}

func (h *LotHandler) List(c *gin.Context) {
	// Parse pagination from query params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filter := domain.LotFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
		Brand:    c.Query("brand"),
		Type:     c.Query("type"),
	}

	lots, total, err := h.service.ListLots(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch lots"})
		return
	}

	// Return empty array instead of null if no lots found (better for frontend)
	if lots == nil {
		lots = []domain.LotResponse{}
	}

	// CRITICAL: Set headers for React Admin
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count") // Expose header to browser

	c.JSON(http.StatusOK, lots)
}