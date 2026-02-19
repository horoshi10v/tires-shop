package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/horoshi10v/tires-shop/internal/domain"
)

type OrderHandler struct {
	service domain.OrderService
}

func NewOrderHandler(service domain.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// Create handles the HTTP request to create an order.
func (h *OrderHandler) Create(c *gin.Context) {
	var req domain.CreateOrderDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request payload",
			"details": err.Error(),
		})
		return
	}

	orderID, err := h.service.CreateOrder(c.Request.Context(), req)
	if err != nil {
		// In a real app, we should check if the error is "not enough stock" to return 409 Conflict.
		// For now, returning 400 is fine.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "order created successfully",
		"order_id": orderID,
	})
}