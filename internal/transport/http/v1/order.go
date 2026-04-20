package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"

	"github.com/horoshi10v/tires-shop/internal/domain"
)

type telegramWebhookMessage struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	Chat      struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	ReplyToMessage *struct {
		MessageID int64 `json:"message_id"`
	} `json:"reply_to_message"`
}

type telegramWebhookUpdate struct {
	Message *telegramWebhookMessage `json:"message"`
}

type OrderHandler struct {
	service         domain.OrderService
	guestOrderGuard *guestOrderGuard
}

func NewOrderHandler(service domain.OrderService) *OrderHandler {
	return &OrderHandler{
		service:         service,
		guestOrderGuard: newGuestOrderGuard(),
	}
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
//	@Success      201    {object}  domain.OrderResponse "Created"
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

	if userID == nil {
		if err := h.guestOrderGuard.Check(c, req); err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
	}

	orderID, err := h.service.CreateOrder(c.Request.Context(), req, userID)
	if err != nil {
		// In a real app, we should check if the error is "not enough stock" to return 409 Conflict.
		// For now, returning 400 is fine.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order created but failed to fetch response"})
		return
	}

	c.JSON(http.StatusCreated, order)
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

func (h *OrderHandler) UpdateItemPrice(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id format"})
		return
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order item id format"})
		return
	}

	var req domain.UpdateOrderItemPriceDTO
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

	if err := h.service.UpdateOrderItemPrice(c.Request.Context(), orderID, itemID, userID, req.Price, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "price updated but failed to fetch order"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// SendMessage sends a direct Telegram bot message to the order customer.
//
//	@Summary      Send Order Message
//	@Description  Send a Telegram bot message to the customer tied to the order via customer_telegram_id.
//	@Tags         orders-staff
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id    path      string                      true  "Order ID"
//	@Param        data  body      domain.SendOrderMessageDTO  true  "Message payload"
//	@Success      200   {object}  map[string]string "OK"
//	@Failure      400   {object}  map[string]string "Bad Request"
//	@Failure      500   {object}  map[string]string "Internal Server Error"
//	@Router       /staff/orders/{id}/message [post]
func (h *OrderHandler) SendMessage(c *gin.Context) {
	orderIDParam := c.Param("id")
	orderID, err := uuid.Parse(orderIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id format"})
		return
	}

	var req domain.SendOrderMessageDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SendOrderMessage(c.Request.Context(), orderID, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "bot message sent"})
}

// ListMessages returns the message history for a specific order.
//
//	@Summary      List Order Messages
//	@Description  Returns inbound and outbound Telegram messages linked to the order.
//	@Tags         orders-staff
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id    path      string  true  "Order ID"
//	@Success      200   {array}   domain.OrderMessage
//	@Failure      400   {object}  map[string]string "Bad Request"
//	@Failure      500   {object}  map[string]string "Internal Server Error"
//	@Router       /staff/orders/{id}/messages [get]
func (h *OrderHandler) ListMessages(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id format"})
		return
	}

	messages, err := h.service.ListOrderMessages(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if messages == nil {
		messages = []domain.OrderMessage{}
	}

	c.JSON(http.StatusOK, messages)
}

func (h *OrderHandler) HandleClientBotWebhook(c *gin.Context) {
	var update telegramWebhookUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
		return
	}

	if update.Message == nil || update.Message.Chat.ID == 0 || update.Message.Text == "" {
		c.JSON(http.StatusOK, gin.H{"message": "ignored"})
		return
	}

	var replyToMessageID *int64
	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.MessageID != 0 {
		replyID := update.Message.ReplyToMessage.MessageID
		replyToMessageID = &replyID
	}

	if err := h.service.ProcessInboundMessage(c.Request.Context(), domain.InboundOrderMessageDTO{
		CustomerTelegramID:       update.Message.Chat.ID,
		TelegramMessageID:        update.Message.MessageID,
		ReplyToTelegramMessageID: replyToMessageID,
		MessageText:              update.Message.Text,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "accepted"})
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
