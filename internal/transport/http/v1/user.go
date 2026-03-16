package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type UserHandler struct {
	service domain.UserService
}

func NewUserHandler(service domain.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// ListUsers retrieves a paginated list of users.
//
//	@Summary      List Users
//	@Description  Get all users with optional filtering.
//	@Tags         users-admin
//	@Produce      json
//	@Security     RoleAuth
//	@Param        page      query     int     false  "Page number" default(1)
//	@Param        page_size query     int     false  "Items per page" default(20)
//	@Param        search    query     string  false  "Search by username, phone, name"
//	@Param        role      query     string  false  "Filter by role (BUYER, STAFF, ADMIN)"
//	@Success      200       {array}   domain.User
//	@Router       /admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	role := c.Query("role")

	filter := domain.UserFilter{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		Role:     domain.UserRole(role),
	}

	users, total, err := h.service.ListUsers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to limit users"})
		return
	}

	if users == nil {
		users = []domain.User{}
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count")
	c.JSON(http.StatusOK, users)
}

// AddWorker creates or promotes a user to worker status.
//
//	@Summary      Add Worker
//	@Description  Create new worker by Telegram ID or promote existing user by Username/Phone.
//	@Tags         users-admin
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        data  body      domain.CreateWorkerDTO  true  "Worker details"
//	@Success      201   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]string
//	@Router       /admin/users [post]
func (h *UserHandler) AddWorker(c *gin.Context) {
	var req domain.CreateWorkerDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	id, err := h.service.AddWorker(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "worker added successfully",
		"user_id": id,
	})
}

// UpdateRole changes a user's role.
//
//	@Summary      Update User Role
//	@Description  Promote or demote a user (e.g. BUYER -> STAFF).
//	@Tags         users-admin
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id    path      string                  true  "User ID"
//	@Param        data  body      domain.UpdateUserRoleDTO true  "New Role"
//	@Success      200   {object}  map[string]string
//	@Router       /admin/users/{id}/role [put]
func (h *UserHandler) UpdateRole(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req domain.UpdateUserRoleDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	if err := h.service.UpdateUserRole(c.Request.Context(), userID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user role updated"})
}

// Delete removes a user from the system.
//
//	@Summary      Delete User
//	@Description  Soft delete a user account.
//	@Tags         users-admin
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id    path      string  true  "User ID"
//	@Success      200   {object}  map[string]string
//	@Router       /admin/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
