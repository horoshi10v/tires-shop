package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

// ListPublic retrieves a paginated and sorted list of public lots.
//
//	@Summary      List available lots (Public)
//	@Description  Get active lots with server-side filtering, sorting, and pagination.
//	@Tags         lots-public
//	@Produce      json
//	@Param        page          query     int     false  "Page number" default(1)
//	@Param        page_size     query     int     false  "Items per page" default(12)
//	@Param        sort_by       query     string  false  "Sort field: price, created_at, stock, popularity"
//	@Param        sort_order    query     string  false  "Sort order: asc or desc"
//	@Param        search        query     string  false  "Search by brand or model"
//	@Param        brand         query     string  false  "Filter by brand name"
//	@Param        type          query     string  false  "Filter by type (TIRE, RIM, ACCESSORY)"
//	@Param        width         query     int     false  "Filter by width (mm)"
//	@Param        profile       query     int     false  "Filter by profile (%)"
//	@Param        diameter      query     int     false  "Filter by diameter (R)"
//	@Param        season        query     string  false  "Filter by season"
//	@Param        model         query     string  false  "Filter by model name"
//	@Param        condition     query     string  false  "Filter by condition (NEW/USED)"
//	@Param        is_run_flat   query     bool    false  "Filter by run flat parameter"
//	@Param        is_spiked     query     bool    false  "Filter by spiked parameter"
//	@Param        anti_puncture query     bool    false  "Filter by anti puncture parameter"
//	@Param        sell_price    query     number  false  "Filter by exact sell price"
//	@Param        current_quantity query int     false  "Filter by exact quantity"
//	@Success      200  {object}  domain.PaginatedLotPublicResponse
//	@Failure      400  {object}  map[string]string
//	@Router       /lots [get]
func (h *LotHandler) ListPublic(c *gin.Context) {
	filter, err := buildPublicLotFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query params", "details": err.Error()})
		return
	}

	lots, total, err := h.service.ListPublicLots(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch lots"})
		return
	}

	if lots == nil {
		lots = []domain.LotPublicResponse{}
	}

	c.JSON(http.StatusOK, domain.PaginatedLotPublicResponse{
		Items:    lots,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Total:    total,
		HasNext:  int64(filter.Page*filter.PageSize) < total,
	})
}

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

func buildPublicLotFilter(c *gin.Context) (domain.LotFilter, error) {
	filter := buildLotFilter(c)

	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PageSize == 0 {
		filter.PageSize = 12
	}
	if filter.Page < 1 {
		return domain.LotFilter{}, fmt.Errorf("page must be greater than or equal to 1")
	}
	if filter.PageSize < 1 || filter.PageSize > 50 {
		return domain.LotFilter{}, fmt.Errorf("page_size must be between 1 and 50")
	}

	filter.SortBy = strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_by", "created_at")))
	filter.SortOrder = strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_order", "desc")))

	allowedSortBy := map[string]bool{
		"price":      true,
		"created_at": true,
		"stock":      true,
		"popularity": true,
	}
	if !allowedSortBy[filter.SortBy] {
		return domain.LotFilter{}, fmt.Errorf("sort_by must be one of: price, created_at, stock, popularity")
	}
	if filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		return domain.LotFilter{}, fmt.Errorf("sort_order must be one of: asc, desc")
	}

	return filter, nil
}

func buildLotFilter(c *gin.Context) domain.LotFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	width, _ := strconv.Atoi(c.Query("width"))
	profile, _ := strconv.Atoi(c.Query("profile"))
	diameter, _ := strconv.Atoi(c.Query("diameter"))
	productionYear, _ := strconv.Atoi(c.Query("production_year"))
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
		SortBy:            c.Query("sort_by"),
		SortOrder:         c.Query("sort_order"),
		Status:            c.Query("status"),
		Brand:             c.Query("brand"),
		Type:              c.Query("type"),
		Search:            c.Query("search"),
		Width:             width,
		Profile:           profile,
		Diameter:          diameter,
		ProductionYear:    productionYear,
		CountryOfOrigin:   c.Query("country_of_origin"),
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

	c.Data(http.StatusOK, "image/png", pngBytes)
}
