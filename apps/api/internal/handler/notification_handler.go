package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notifSvc service.NotificationService
}

func NewNotificationHandler(notifSvc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifSvc: notifSvc}
}

func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	notifications, total, unreadCount, err := h.notifSvc.GetUserNotifications(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get notifications error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notifications"})
		return
	}

	totalPages := total / int32(limit)
	if total%int32(limit) > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.NotificationListResponse{
			Notifications: mapper.ToNotificationResponses(notifications),
			Total:         total,
			UnreadCount:   unreadCount,
			Page:          int32(page),
			Limit:         int32(limit),
			TotalPages:    totalPages,
		},
	})
}

func (h *NotificationHandler) GetUnreadNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	notifications, count, err := h.notifSvc.GetUnreadNotifications(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get unread notifications error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notifications"})
		return
	}

	totalPages := count / int32(limit)
	if count%int32(limit) > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.NotificationListResponse{
			Notifications: mapper.ToNotificationResponses(notifications),
			Total:         count,
			UnreadCount:   count,
			Page:          int32(page),
			Limit:         int32(limit),
			TotalPages:    totalPages,
		},
	})
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	_, _, unreadCount, err := h.notifSvc.GetUserNotifications(c.Request.Context(), userID, 1, 0)
	if err != nil {
		log.Printf("get unread count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"unread_count": unreadCount}})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	notifIDStr := c.Param("notificationId")
	notifID, err := strconv.Atoi(notifIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	if err := h.notifSvc.MarkNotificationAsRead(c.Request.Context(), userID, int32(notifID)); err != nil {
		switch {
		case errors.Is(err, service.ErrNotificationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		default:
			log.Printf("mark as read error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notification"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	if err := h.notifSvc.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		log.Printf("mark all as read error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}

func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	notifIDStr := c.Param("notificationId")
	notifID, err := strconv.Atoi(notifIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	if err := h.notifSvc.DeleteNotification(c.Request.Context(), userID, int32(notifID)); err != nil {
		switch {
		case errors.Is(err, service.ErrNotificationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		default:
			log.Printf("delete notification error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete notification"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification deleted"})
}

func (h *NotificationHandler) DeleteAllNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	if err := h.notifSvc.DeleteAllNotifications(c.Request.Context(), userID); err != nil {
		log.Printf("delete all notifications error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all notifications deleted"})
}
