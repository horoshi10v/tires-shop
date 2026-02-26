package v1

import (
	"net/http"

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
