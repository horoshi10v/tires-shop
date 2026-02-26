package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type WarehouseHandler struct {
	service domain.WarehouseService
}

func NewWarehouseHandler(service domain.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{service: service}
}

// Create handles the creation of a new warehouse.
//
//	@Summary      Create Warehouse
//	@Tags         warehouses
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        data  body      domain.CreateWarehouseDTO  true  "Warehouse details"
//	@Success      201   {object}  map[string]interface{}
//	@Router       /admin/warehouses [post]
func (h *WarehouseHandler) Create(c *gin.Context) {
	var req domain.CreateWarehouseDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.CreateWarehouse(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "warehouse created", "warehouse_id": id})
}

// List retrieves all warehouses.
//
//	@Summary      List Warehouses
//	@Tags         warehouses
//	@Produce      json
//	@Security     RoleAuth
//	@Success      200   {array}   domain.Warehouse
//	@Router       /staff/warehouses [get]
func (h *WarehouseHandler) List(c *gin.Context) {
	warehouses, err := h.service.ListWarehouses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch warehouses"})
		return
	}
	c.JSON(http.StatusOK, warehouses)
}

// Update modifies an existing warehouse.
//
//	@Summary      Update Warehouse
//	@Tags         warehouses
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id    path      string                     true  "Warehouse ID"
//	@Param        data  body      domain.UpdateWarehouseDTO  true  "Updated details"
//	@Success      200   {object}  map[string]string
//	@Router       /admin/warehouses/{id} [put]
func (h *WarehouseHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var req domain.UpdateWarehouseDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateWarehouse(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "warehouse updated"})
}

// Delete performs a soft delete on a warehouse if it has no active stock.
//
//	@Summary      Delete Warehouse
//	@Tags         warehouses
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id    path      string  true  "Warehouse ID"
//	@Success      200   {object}  map[string]string
//	@Router       /admin/warehouses/{id} [delete]
func (h *WarehouseHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	if err := h.service.DeleteWarehouse(c.Request.Context(), id); err != nil {
		// HTTP 409 Conflict is the correct status code when a deletion fails due to relation constraints
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "warehouse deleted"})
}
