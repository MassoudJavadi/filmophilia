package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	activitySvc service.ActivityService
}

func NewActivityHandler(activitySvc service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activitySvc: activitySvc}
}

// CreateActivity godoc
// @Summary Create an activity
// @Description Log a user activity (rating, comment, watchlist action, etc.)
// @Tags activities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateActivityRequest true "Activity details"
// @Success 201 {object} map[string]string "Activity created"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /activities [post]
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	var req dto.CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.activitySvc.CreateActivity(c.Request.Context(), userID, req.Action, req.EntityType, req.EntityID, req.Metadata); err != nil {
		log.Printf("create activity error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create activity"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "activity created"})
}

// GetMyActivities godoc
// @Summary Get my activities
// @Description Get the authenticated user's activity history
// @Tags activities
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "Activity feed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /me/activities [get]
func (h *ActivityHandler) GetMyActivities(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	activities, total, err := h.activitySvc.GetUserActivities(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get my activities error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch activities"})
		return
	}

	totalPages := total / int32(limit)
	if total%int32(limit) > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.ActivityFeedResponse{
			Activities: mapper.ToActivityResponses(activities),
			Total:      total,
			Page:       int32(page),
			Limit:      int32(limit),
			TotalPages: totalPages,
		},
	})
}

// GetUserActivities godoc
// @Summary Get user activities
// @Description Get a specific user's public activity history
// @Tags activities
// @Produce json
// @Param userId path int true "User ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "Activity feed"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{userId}/activities [get]
func (h *ActivityHandler) GetUserActivities(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	activities, total, err := h.activitySvc.GetUserActivities(c.Request.Context(), int32(userID), int32(limit), int32(offset))
	if err != nil {
		log.Printf("get user activities error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch activities"})
		return
	}

	totalPages := total / int32(limit)
	if total%int32(limit) > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.ActivityFeedResponse{
			Activities: mapper.ToActivityResponses(activities),
			Total:      total,
			Page:       int32(page),
			Limit:      int32(limit),
			TotalPages: totalPages,
		},
	})
}

// GetFollowingFeed godoc
// @Summary Get following feed
// @Description Get activities from users the authenticated user follows
// @Tags activities
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "Activity feed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /me/feed [get]
func (h *ActivityHandler) GetFollowingFeed(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	activities, total, err := h.activitySvc.GetFollowingActivities(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get following feed error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch feed"})
		return
	}

	totalPages := total / int32(limit)
	if total%int32(limit) > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.ActivityFeedResponse{
			Activities: mapper.ToFollowingActivityResponses(activities),
			Total:      total,
			Page:       int32(page),
			Limit:      int32(limit),
			TotalPages: totalPages,
		},
	})
}

// GetMovieActivities godoc
// @Summary Get movie activities
// @Description Get all activities related to a specific movie
// @Tags activities
// @Produce json
// @Param movieId path int true "Movie ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "List of activities"
// @Failure 400 {object} map[string]string "Invalid movie ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /movies/{movieId}/activities [get]
func (h *ActivityHandler) GetMovieActivities(c *gin.Context) {
	movieIDStr := c.Param("movieId")
	movieID, err := strconv.ParseInt(movieIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	// Entity type 1 = MOVIE (based on enum in schema)
	activities, err := h.activitySvc.GetActivitiesByEntity(c.Request.Context(), 1, movieID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get movie activities error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToEntityActivityResponses(activities)})
}
