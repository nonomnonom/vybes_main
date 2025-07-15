package http

import (
	"net/http"
	"strconv"
	"vybes/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationHandler handles HTTP requests for notifications.
type NotificationHandler struct {
	notificationService service.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications is the handler for fetching a user's notifications.
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))

	notifications, err := h.notificationService.GetNotifications(c.Request.Context(), userID.(primitive.ObjectID).Hex(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}
	c.JSON(http.StatusOK, notifications)
}

// MarkAsRead is the handler for marking notifications as read.
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var request struct {
		NotificationIDs []string `json:"notificationIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	modifiedCount, err := h.notificationService.MarkNotificationsAsRead(c.Request.Context(), userID.(primitive.ObjectID).Hex(), request.NotificationIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"modifiedCount": modifiedCount})
}
