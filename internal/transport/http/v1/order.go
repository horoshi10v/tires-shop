package v1

import (
	"net/http"
	"strconv"

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

	var userID *uuid.UUID
	if val, exists := c.Get("userID"); exists {
		if id, ok := val.(uuid.UUID); ok {
			userID = &id
		}
	}

	orderID, err := h.service.CreateOrder(c.Request.Context(), req, userID)
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
//	@Summary      Update Order Status
//	@Description  Change the status of an existing order (e.g., NEW, PREPAYMENT, DONE, CANCELLED).
//	@Tags         orders
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id    path      string                       true  "Order ID"
//	@Param        data  body      domain.UpdateOrderStatusDTO  true  "New status and comment"
//	@Success      200   {object}  map[string]string "OK"
//	@Failure      400   {object}  map[string]string "Bad Request"
//	@Failure      401   {object}  map[string]string "Unauthorized"
//	@Failure      403   {object}  map[string]string "Forbidden"
//	@Failure      500   {object}  map[string]string "Internal Server Error"
//	@Router       /staff/orders/{id}/status [put]
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

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user identification missing"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	if err := h.service.UpdateOrderStatus(c.Request.Context(), orderID, req.Status, userID, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated and logged"})
}

// List handles listing orders with filters.
//
//	@Summary      List Orders
//	@Description  Get a paginated list of orders, filterable by status or customer.
//	@Tags         orders-staff
//	@Produce      json
//	@Security     RoleAuth
//	@Param        page      query     int     false  "Page number" default(1)
//	@Param        page_size query     int     false  "Items per page" default(10)
//	@Param        status    query     string  false  "Filter by status"
//	@Param        customer  query     string  false  "Search by customer name or phone"
//	@Success      200       {array}   domain.OrderResponse
//	@Failure      500       {object}  map[string]string "Internal Server Error"
//	@Router       /staff/orders [get]
func (h *OrderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	customer := c.Query("customer")

	filter := domain.OrderFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Customer: customer,
	}

	orders, total, err := h.service.ListOrders(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	if orders == nil {
		orders = []domain.OrderResponse{}
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count")
	c.JSON(http.StatusOK, orders)
}

// ListMyOrders handles listing orders for the authenticated user.
//
//	@Summary      My Orders History
//	@Description  Get a paginated list of orders placed by the current user.
//	@Tags         orders
//	@Produce      json
//	@Security     RoleAuth
//	@Param        page      query     int     false  "Page number" default(1)
//	@Param        page_size query     int     false  "Items per page" default(10)
//	@Param        status    query     string  false  "Filter by status"
//	@Success      200       {array}   domain.OrderResponse
//	@Failure      500       {object}  map[string]string "Internal Server Error"
//	@Router       /orders [get]
func (h *OrderHandler) ListMyOrders(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user identification missing"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	filter := domain.OrderFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	}

	orders, total, err := h.service.ListMyOrders(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch your orders"})
		return
	}

	if orders == nil {
		orders = []domain.OrderResponse{}
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count")
	c.JSON(http.StatusOK, orders)
}
