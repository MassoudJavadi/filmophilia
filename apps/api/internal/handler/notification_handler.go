package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/response"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notifSvc service.NotificationService
}

func NewNotificationHandler(notifSvc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifSvc: notifSvc}
}

// GetMyNotifications godoc
// @Summary Get my notifications
// @Description Get all notifications for the authenticated user
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=dto.NotificationListResponse} "Notifications"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /me/notifications [get]
func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	notifications, total, unreadCount, err := h.notifSvc.GetUserNotifications(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get notifications error: %v", err)
		response.Error(c, http.StatusInternalServerError, "NOTIFICATIONS_FETCH_FAILED", "failed to fetch notifications")
		return
	}

	totalPages := total / int32(limit)
	if total%int32(limit) > 0 {
		totalPages++
	}

	response.OK(c, dto.NotificationListResponse{
		Notifications: mapper.ToNotificationResponses(notifications),
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          int32(page),
		Limit:         int32(limit),
		TotalPages:    totalPages,
	})
}

// GetUnreadNotifications godoc
// @Summary Get unread notifications
// @Description Get only unread notifications for the authenticated user
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=dto.NotificationListResponse} "Unread notifications"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /me/notifications/unread [get]
func (h *NotificationHandler) GetUnreadNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	notifications, count, err := h.notifSvc.GetUnreadNotifications(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get unread notifications error: %v", err)
		response.Error(c, http.StatusInternalServerError, "NOTIFICATIONS_FETCH_FAILED", "failed to fetch notifications")
		return
	}

	totalPages := count / int32(limit)
	if count%int32(limit) > 0 {
		totalPages++
	}

	response.OK(c, dto.NotificationListResponse{
		Notifications: mapper.ToNotificationResponses(notifications),
		Total:         count,
		UnreadCount:   count,
		Page:          int32(page),
		Limit:         int32(limit),
		TotalPages:    totalPages,
	})
}

// GetUnreadCount godoc
// @Summary Get unread notification count
// @Description Get the count of unread notifications
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.UnreadCountData} "Unread count"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /me/notifications/unread/count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	_, _, unreadCount, err := h.notifSvc.GetUserNotifications(c.Request.Context(), userID, 1, 0)
	if err != nil {
		log.Printf("get unread count error: %v", err)
		response.Error(c, http.StatusInternalServerError, "COUNT_FETCH_FAILED", "failed to fetch count")
		return
	}

	response.OK(c, dto.UnreadCountData{UnreadCount: unreadCount})
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param notificationId path int true "Notification ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageData} "Marked as read"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid notification ID"
// @Failure 403 {object} dto.ForbiddenErrorResponse "Unauthorized"
// @Failure 404 {object} dto.NotFoundErrorResponse "Notification not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /notifications/{notificationId}/read [patch]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	notifIDStr := c.Param("notificationId")
	notifID, err := strconv.ParseInt(notifIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_NOTIFICATION_ID", "invalid notification id")
		return
	}

	if err := h.notifSvc.MarkNotificationAsRead(c.Request.Context(), userID, notifID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotificationNotFound):
			response.Error(c, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "notification not found")
		case errors.Is(err, service.ErrUnauthorized):
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "unauthorized")
		default:
			log.Printf("mark as read error: %v", err)
			response.Error(c, http.StatusInternalServerError, "NOTIFICATION_UPDATE_FAILED", "failed to update notification")
		}
		return
	}

	response.Message(c, http.StatusOK, "notification marked as read")
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all notifications for the authenticated user as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageData} "All marked as read"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /notifications/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	if err := h.notifSvc.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		log.Printf("mark all as read error: %v", err)
		response.Error(c, http.StatusInternalServerError, "NOTIFICATIONS_UPDATE_FAILED", "failed to update notifications")
		return
	}

	response.Message(c, http.StatusOK, "all notifications marked as read")
}

// DeleteNotification godoc
// @Summary Delete a notification
// @Description Delete a specific notification
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param notificationId path int true "Notification ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageData} "Notification deleted"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid notification ID"
// @Failure 403 {object} dto.ForbiddenErrorResponse "Unauthorized"
// @Failure 404 {object} dto.NotFoundErrorResponse "Notification not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /notifications/{notificationId} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	notifIDStr := c.Param("notificationId")
	notifID, err := strconv.ParseInt(notifIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_NOTIFICATION_ID", "invalid notification id")
		return
	}

	if err := h.notifSvc.DeleteNotification(c.Request.Context(), userID, notifID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotificationNotFound):
			response.Error(c, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "notification not found")
		case errors.Is(err, service.ErrUnauthorized):
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "unauthorized")
		default:
			log.Printf("delete notification error: %v", err)
			response.Error(c, http.StatusInternalServerError, "NOTIFICATION_DELETE_FAILED", "failed to delete notification")
		}
		return
	}

	response.Message(c, http.StatusOK, "notification deleted")
}

// DeleteAllNotifications godoc
// @Summary Delete all notifications
// @Description Delete all notifications for the authenticated user
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageData} "All notifications deleted"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /notifications [delete]
func (h *NotificationHandler) DeleteAllNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	if err := h.notifSvc.DeleteAllNotifications(c.Request.Context(), userID); err != nil {
		log.Printf("delete all notifications error: %v", err)
		response.Error(c, http.StatusInternalServerError, "NOTIFICATIONS_DELETE_FAILED", "failed to delete notifications")
		return
	}

	response.Message(c, http.StatusOK, "all notifications deleted")
}
