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

// Create handles the HTTP request to create a new lot.
//
//	@Summary      Create a new lot
//	@Description  Add a new lot to the inventory.
//	@Tags         lots-staff
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        lot  body      domain.CreateLotDTO  true  "Lot details"
//	@Success      201  {object}  map[string]interface{}
//	@Router       /staff/lots [post]
func (h *LotHandler) Create(c *gin.Context) {
	var req domain.CreateLotDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	id, err := h.service.CreateLot(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "lot created", "lot_id": id})
}

// ListPublic retrieves a list of lots for buyers (hides sensitive data).
//
//	@Summary      List available lots (Public)
//	@Description  Get active lots. Purchase price and archive data are hidden.
//	@Tags         lots-public
//	@Produce      json
//	@Param        page      query     int     false  "Page number" default(1)
//	@Param        page_size query     int     false  "Items per page" default(10)
//	@Param        brand     query     string  false  "Filter by brand name"
//	@Success      200       {array}   domain.LotPublicResponse
//	@Router       /lots [get]
func (h *LotHandler) ListPublic(c *gin.Context) {
	filter := buildLotFilter(c)

	lots, total, err := h.service.ListPublicLots(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch lots"})
		return
	}

	if lots == nil {
		lots = []domain.LotPublicResponse{}
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count")
	c.JSON(http.StatusOK, lots)
}

// ListInternal retrieves a full list of lots for staff/admin.
//
//	@Summary      List all lots (Internal)
//	@Description  Get lots including sensitive financial data and archives.
//	@Tags         lots-staff
//	@Produce      json
//	@Security     RoleAuth
//	@Param        page      query     int     false  "Page number" default(1)
//	@Param        page_size query     int     false  "Items per page" default(10)
//	@Param        brand     query     string  false  "Filter by brand name"
//	@Param        status    query     string  false  "Filter by status"
//	@Success      200       {array}   domain.LotInternalResponse
//	@Router       /staff/lots [get]
func (h *LotHandler) ListInternal(c *gin.Context) {
	filter := buildLotFilter(c)

	lots, total, err := h.service.ListInternalLots(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch lots"})
		return
	}

	if lots == nil {
		lots = []domain.LotInternalResponse{}
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count")
	c.JSON(http.StatusOK, lots)
}

// Helper to extract query params
func buildLotFilter(c *gin.Context) domain.LotFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	return domain.LotFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
		Brand:    c.Query("brand"),
		Type:     c.Query("type"),
	}
}
