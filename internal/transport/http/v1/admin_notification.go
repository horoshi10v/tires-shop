package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type AdminNotificationHandler struct {
	service domain.AdminNotificationService
}

func NewAdminNotificationHandler(service domain.AdminNotificationService) *AdminNotificationHandler {
	return &AdminNotificationHandler{service: service}
}

// List returns admin notifications.
//
//	@Summary      List Admin Notifications
//	@Description  Returns a paginated list of admin notifications.
//	@Tags         notifications-admin
//	@Produce      json
//	@Security     RoleAuth
//	@Param        page       query     int     false  "Page number" default(1)
//	@Param        page_size  query     int     false  "Items per page" default(20)
//	@Param        type       query     string  false  "Filter by notification type"
//	@Param        is_read    query     bool    false  "Filter by read state"
//	@Success      200        {array}   domain.AdminNotification
//	@Router       /admin/notifications [get]
func (h *AdminNotificationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var isRead *bool
	if rawIsRead := c.Query("is_read"); rawIsRead != "" {
		parsed, err := strconv.ParseBool(rawIsRead)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid is_read value"})
			return
		}
		isRead = &parsed
	}

	notifications, total, err := h.service.List(c.Request.Context(), domain.AdminNotificationFilter{
		Page:     page,
		PageSize: pageSize,
		Type:     domain.AdminNotificationType(c.Query("type")),
		IsRead:   isRead,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notifications"})
		return
	}

	if notifications == nil {
		notifications = []domain.AdminNotification{}
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count")
	c.JSON(http.StatusOK, notifications)
}

// MarkRead marks a notification as read.
//
//	@Summary      Mark Notification Read
//	@Description  Marks an admin notification as read.
//	@Tags         notifications-admin
//	@Produce      json
//	@Security     RoleAuth
//	@Param        id   path      string  true  "Notification ID"
//	@Success      200  {object}  map[string]string
//	@Router       /admin/notifications/{id}/read [post]
func (h *AdminNotificationHandler) MarkRead(c *gin.Context) {
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	if err := h.service.MarkRead(c.Request.Context(), notificationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}
