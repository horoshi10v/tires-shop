package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"

	"github.com/horoshi10v/tires-shop/internal/domain"
)

type OrderHandler struct {
	service domain.OrderService
}

func NewOrderHandler(service domain.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// Create handles the HTTP request to create an order.
//
//	@Summary      Create a new order
//	@Description  Place a new order with customer details and tire items.
//	@Tags         orders
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        order  body      domain.CreateOrderDTO  true  "Order details"
//	@Success      201    {object}  map[string]interface{} "Created"
//	@Failure      400    {object}  map[string]string "Bad Request"
//	@Failure      401    {object}  map[string]string "Unauthorized"
//	@Failure      403    {object}  map[string]string "Forbidden"
//	@Failure      500    {object}  map[string]string "Internal Server Error"
//	@Router       /orders [post]
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

// UpdateStatus handles the HTTP request to change an order's status.
//
//	@Summary      Update order status
//	@Description  Change the status of an existing order (e.g., NEW, PREPAYMENT, DONE, CANCELLED).
//	@Tags         orders
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id     path      string                 true  "Order ID"
//	@Param        status body      domain.UpdateOrderStatusDTO true "New status"
//	@Success      200    {object}  map[string]string "OK"
//	@Failure      400    {object}  map[string]string "Bad Request"
//	@Failure      401    {object}  map[string]string "Unauthorized"
//	@Failure      403    {object}  map[string]string "Forbidden"
//	@Failure      500    {object}  map[string]string "Internal Server Error"
//	@Router       /orders/{id}/status [put]
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	orderIDParam := c.Param("id")
	orderID, err := uuid.Parse(orderIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id format"})
		return
	}

	var req domain.UpdateOrderStatusDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateOrderStatus(c.Request.Context(), orderID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated"})
}
