package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type TransferHandler struct {
	service domain.TransferService
}

func NewTransferHandler(service domain.TransferService) *TransferHandler {
	return &TransferHandler{service: service}
}

// Create initiates a new stock transfer.
//
//	@Summary      Create Transfer
//	@Description  Moves stock from one warehouse to another (Status: IN_TRANSIT).
//	@Tags         transfers
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        data  body      domain.CreateTransferDTO  true  "Transfer details"
//	@Success      201   {object}  map[string]interface{}
//	@Router       /staff/transfers [post]
func (h *TransferHandler) Create(c *gin.Context) {
	var req domain.CreateTransferDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	transferID, err := h.service.CreateTransfer(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "transfer created", "transfer_id": transferID})
}

// Accept completes the transfer and creates new lots at destination.
//
//	@Summary      Accept Transfer
//	@Description  Receives the stock at the destination warehouse (Status: ACCEPTED).
//	@Tags         transfers
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id    path      string  true  "Transfer ID"
//	@Success      200   {object}  map[string]string
//	@Router       /staff/transfers/{id}/accept [post]
func (h *TransferHandler) Accept(c *gin.Context) {
	transferID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer id"})
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	if err := h.service.AcceptTransfer(c.Request.Context(), transferID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer accepted successfully"})
}

// Cancel cancels a pending transfer and returns stock to source.
//
//	@Summary      Cancel Transfer
//	@Description  Cancels an IN_TRANSIT transfer and refunds stock.
//	@Tags         transfers
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id    path      string  true  "Transfer ID"
//	@Success      200   {object}  map[string]string
//	@Router       /staff/transfers/{id}/cancel [post]
func (h *TransferHandler) Cancel(c *gin.Context) {
	transferID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer id"})
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	if err := h.service.CancelTransfer(c.Request.Context(), transferID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer cancelled successfully"})
}

// List retrieves a list of transfers.
//
//	@Summary      List Transfers
//	@Description  Get paginated list of transfers.
//	@Tags         transfers
//	@Produce      json
//	@Security     RoleAuth
//	@Param        page              query     int     false  "Page number" default(1)
//	@Param        page_size         query     int     false  "Items per page" default(10)
//	@Param        status            query     string  false  "Filter by status"
//	@Param        from_warehouse_id query     string  false  "Filter by source warehouse"
//	@Param        to_warehouse_id   query     string  false  "Filter by destination warehouse"
//	@Success      200               {array}   domain.TransferResponse
//	@Router       /staff/transfers [get]
func (h *TransferHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filter := domain.TransferFilter{
		Page:            page,
		PageSize:        pageSize,
		Status:          c.Query("status"),
		FromWarehouseID: c.Query("from_warehouse_id"),
		ToWarehouseID:   c.Query("to_warehouse_id"),
	}

	transfers, total, err := h.service.ListTransfers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transfers"})
		return
	}

	if transfers == nil {
		transfers = []domain.TransferResponse{}
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count")
	c.JSON(http.StatusOK, transfers)
}

// GetByID retrieves a single transfer details.
//
//	@Summary      Get Transfer Details
//	@Description  Get transfer by ID.
//	@Tags         transfers
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id    path      string  true  "Transfer ID"
//	@Success      200   {object}  domain.TransferResponse
//	@Router       /staff/transfers/{id} [get]
func (h *TransferHandler) GetByID(c *gin.Context) {
	transferID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer id"})
		return
	}

	transfer, err := h.service.GetTransfer(c.Request.Context(), transferID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transfer not found"})
		return
	}

	c.JSON(http.StatusOK, transfer)
}
