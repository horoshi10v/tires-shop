package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// Update handles updating an existing lot.
//
//	@Summary      Update a lot
//	@Description  Update details of an existing lot.
//	@Tags         lots-staff
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id   path      string               true  "Lot ID"
//	@Param        lot  body      domain.UpdateLotDTO  true  "Lot update details"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /staff/lots/{id} [put]
func (h *LotHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	lotID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lot id format"})
		return
	}

	var req domain.UpdateLotDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	if err := h.service.UpdateLot(c.Request.Context(), lotID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update lot", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "lot updated successfully"})
}

// Delete handles deleting a lot.
//
//	@Summary      Delete a lot
//	@Description  Soft delete a lot.
//	@Tags         lots-staff
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id   path      string  true  "Lot ID"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /staff/lots/{id} [delete]
func (h *LotHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	lotID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lot id format"})
		return
	}

	if err := h.service.DeleteLot(c.Request.Context(), lotID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete lot", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "lot deleted successfully"})
}

// ListPublic retrieves a list of lots for buyers (hides sensitive data).
//
//	@Summary      List available lots (Public)
//	@Description  Get active lots. Purchase price and archive data are hidden.
//	@Tags         lots-public
//	@Produce      json
//	@Param        page      query     int     false  "Page number" default(1)
//	@Param        page_size query     int     false  "Items per page" default(10)
//	@Param        search    query     string  false  "Search by brand or model"
//	@Param        brand     query     string  false  "Filter by brand name"
//	@Param        type      query     string  false  "Filter by type (TIRE, RIM)"
//	@Param        width     query     int     false  "Filter by width (mm)"
//	@Param        profile   query     int     false  "Filter by profile (%)"
//	@Param        diameter  query     int     false  "Filter by diameter (R)"
//	@Param        season    query     string  false  "Filter by season"
//	@Param        model     query     string  false  "Filter by model name"
//	@Param        condition query     string  false  "Filter by condition (NEW/USED)"
//	@Param        is_run_flat      query     bool    false  "Filter by run flat parameter"
//	@Param        is_spiked        query     bool    false  "Filter by spiked parameter"
//	@Param        anti_puncture    query     bool    false  "Filter by anti puncture parameter"
//	@Param        sell_price       query     number  false  "Filter by exact sell price"
//	@Param        current_quantity query     int     false  "Filter by exact quantity"
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
//	@Param        search    query     string  false  "Search by brand or model"
//	@Param        brand     query     string  false  "Filter by brand name"
//	@Param        status    query     string  false  "Filter by status"
//	@Param        type      query     string  false  "Filter by type (TIRE, RIM)"
//	@Param        width     query     int     false  "Filter by width (mm)"
//	@Param        profile   query     int     false  "Filter by profile (%)"
//	@Param        diameter  query     int     false  "Filter by diameter (R)"
//	@Param        season    query     string  false  "Filter by season"
//	@Param        model     query     string  false  "Filter by model name"
//	@Param        condition query     string  false  "Filter by condition (NEW/USED)"
//	@Param        is_run_flat      query     bool    false  "Filter by run flat parameter"
//	@Param        is_spiked        query     bool    false  "Filter by spiked parameter"
//	@Param        anti_puncture    query     bool    false  "Filter by anti puncture parameter"
//	@Param        sell_price       query     number  false  "Filter by exact sell price"
//	@Param        current_quantity query     int     false  "Filter by exact quantity"
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

	width, _ := strconv.Atoi(c.Query("width"))
	profile, _ := strconv.Atoi(c.Query("profile"))
	diameter, _ := strconv.Atoi(c.Query("diameter"))
	ringInnerDiameter, _ := strconv.Atoi(c.Query("ring_inner_diameter"))
	ringOuterDiameter, _ := strconv.Atoi(c.Query("ring_outer_diameter"))
	spacerThickness, _ := strconv.Atoi(c.Query("spacer_thickness"))
	packageQuantity, _ := strconv.Atoi(c.Query("package_quantity"))

	var isRunFlat *bool
	if val := c.Query("is_run_flat"); val != "" {
		b, _ := strconv.ParseBool(val)
		isRunFlat = &b
	}

	var isSpiked *bool
	if val := c.Query("is_spiked"); val != "" {
		b, _ := strconv.ParseBool(val)
		isSpiked = &b
	}

	var antiPuncture *bool
	if val := c.Query("anti_puncture"); val != "" {
		b, _ := strconv.ParseBool(val)
		antiPuncture = &b
	}

	var currentQuantity *int
	if val := c.Query("current_quantity"); val != "" {
		i, _ := strconv.Atoi(val)
		currentQuantity = &i
	}

	var sellPrice *float64
	if val := c.Query("sell_price"); val != "" {
		f, _ := strconv.ParseFloat(val, 64)
		sellPrice = &f
	}

	return domain.LotFilter{
		Page:              page,
		PageSize:          pageSize,
		Status:            c.Query("status"),
		Brand:             c.Query("brand"),
		Type:              c.Query("type"),
		Search:            c.Query("search"),
		Width:             width,
		Profile:           profile,
		Diameter:          diameter,
		Season:            c.Query("season"),
		Condition:         c.Query("condition"),
		Model:             c.Query("model"),
		IsRunFlat:         isRunFlat,
		IsSpiked:          isSpiked,
		AntiPuncture:      antiPuncture,
		CurrentQuantity:   currentQuantity,
		SellPrice:         sellPrice,
		AccessoryCategory: c.Query("accessory_category"),
		FastenerType:      c.Query("fastener_type"),
		ThreadSize:        c.Query("thread_size"),
		SeatType:          c.Query("seat_type"),
		RingInnerDiameter: ringInnerDiameter,
		RingOuterDiameter: ringOuterDiameter,
		SpacerType:        c.Query("spacer_type"),
		SpacerThickness:   spacerThickness,
		PackageQuantity:   packageQuantity,
	}
}

// GetQR returns a PNG image of the QR code for a specific lot.
//
//	@Summary      Get Lot QR Code
//	@Description  Generates and returns a PNG image of a QR code containing the Lot ID.
//	@Tags         lots-admin
//	@Produce      image/png
//	@Security     RoleAuth
//	@Param        id   path      string  true  "Lot ID"
//	@Success      200  {file}    file    "PNG Image"
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /staff/lots/{id}/qr [get]
func (h *LotHandler) GetQR(c *gin.Context) {
	idParam := c.Param("id")
	lotID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lot id format"})
		return
	}

	pngBytes, err := h.service.GenerateLotQR(c.Request.Context(), lotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate qr code"})
		return
	}

	// Магия Gin: отдаем сырые байты как картинку
	c.Data(http.StatusOK, "image/png", pngBytes)
}
