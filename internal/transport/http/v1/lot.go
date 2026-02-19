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
//	@Description  Add a new lot to the inventory with details like brand, type, and quantity.
//	@Tags         lots
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        lot  body      domain.CreateLotDTO  true  "Lot details"
//	@Success      201  {object}  map[string]interface{} "Created"
//	@Failure      400  {object}  map[string]string "Bad Request"
//	@Failure      401  {object}  map[string]string "Unauthorized"
//	@Failure      403  {object}  map[string]string "Forbidden"
//	@Failure      500  {object}  map[string]string "Internal Server Error"
//	@Router       /lots [post]
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

// List retrieves a paginated and filtered list of lots.
//
//	@Summary      List all active lots
//	@Description  Get a paginated list of lots with optional filtering by brand or type.
//	@Tags         lots
//	@Accept       json
//	@Produce      json
//	@Param        page      query     int     false  "Page number" default(1)
//	@Param        page_size query     int     false  "Items per page" default(10)
//	@Param        brand     query     string  false  "Filter by brand name (ILIKE)"
//	@Param        status    query     string  false  "Filter by status (e.g. ACTIVE, ARCHIVED)"
//	@Success      200       {array}   domain.LotResponse
//	@Failure      500       {object}  map[string]string "Internal Server Error"
//	@Router       /lots [get]
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
