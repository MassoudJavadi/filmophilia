package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type FollowHandler struct {
	followSvc service.FollowService
}

func NewFollowHandler(followSvc service.FollowService) *FollowHandler {
	return &FollowHandler{followSvc: followSvc}
}

// Follow godoc
// @Summary Follow a user
// @Description Follow another user
// @Tags follows
// @Produce json
// @Security BearerAuth
// @Param userId path int true "User ID to follow"
// @Success 201 {object} map[string]string "Followed successfully"
// @Failure 400 {object} map[string]string "Invalid user ID or cannot follow self"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 409 {object} map[string]string "Already following"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{userId}/follow [post]
func (h *FollowHandler) Follow(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	targetIDStr := c.Param("userId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.followSvc.Follow(c.Request.Context(), userID, int32(targetID)); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotFollowSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
		case errors.Is(err, service.ErrAlreadyFollowing):
			c.JSON(http.StatusConflict, gin.H{"error": "already following this user"})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			log.Printf("follow error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "followed successfully"})
}

// Unfollow godoc
// @Summary Unfollow a user
// @Description Stop following another user
// @Tags follows
// @Produce json
// @Security BearerAuth
// @Param userId path int true "User ID to unfollow"
// @Success 200 {object} map[string]string "Unfollowed successfully"
// @Failure 400 {object} map[string]string "Invalid user ID or cannot unfollow self"
// @Failure 404 {object} map[string]string "Not following this user"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{userId}/follow [delete]
func (h *FollowHandler) Unfollow(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	targetIDStr := c.Param("userId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.followSvc.Unfollow(c.Request.Context(), userID, int32(targetID)); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotFollowSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot unfollow yourself"})
		case errors.Is(err, service.ErrNotFollowing):
			c.JSON(http.StatusNotFound, gin.H{"error": "not following this user"})
		default:
			log.Printf("unfollow error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unfollowed successfully"})
}

// IsFollowing godoc
// @Summary Check if following
// @Description Check if the authenticated user is following another user
// @Tags follows
// @Produce json
// @Security BearerAuth
// @Param userId path int true "User ID to check"
// @Success 200 {object} object "Following status"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{userId}/following/check [get]
func (h *FollowHandler) IsFollowing(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	targetIDStr := c.Param("userId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	isFollowing, err := h.followSvc.IsFollowing(c.Request.Context(), userID, int32(targetID))
	if err != nil {
		log.Printf("is following error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"is_following": isFollowing}})
}

// GetFollowers godoc
// @Summary Get user's followers
// @Description Get list of users following a specific user
// @Tags follows
// @Produce json
// @Param userId path int true "User ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "List of followers"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{userId}/followers [get]
func (h *FollowHandler) GetFollowers(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	followers, err := h.followSvc.GetFollowers(c.Request.Context(), int32(userID), int32(limit), int32(offset))
	if err != nil {
		log.Printf("get followers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToFollowUserResponses(followers)})
}

// GetFollowing godoc
// @Summary Get user's following
// @Description Get list of users that a specific user is following
// @Tags follows
// @Produce json
// @Param userId path int true "User ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "List of following"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{userId}/following [get]
func (h *FollowHandler) GetFollowing(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	following, err := h.followSvc.GetFollowing(c.Request.Context(), int32(userID), int32(limit), int32(offset))
	if err != nil {
		log.Printf("get following error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToFollowingUserResponses(following)})
}

// GetMyFollowers godoc
// @Summary Get my followers
// @Description Get list of users following the authenticated user
// @Tags follows
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "List of followers"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /me/followers [get]
func (h *FollowHandler) GetMyFollowers(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	followers, err := h.followSvc.GetFollowers(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get my followers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToFollowUserResponses(followers)})
}

// GetMyFollowing godoc
// @Summary Get my following
// @Description Get list of users the authenticated user is following
// @Tags follows
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "List of following"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /me/following [get]
func (h *FollowHandler) GetMyFollowing(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	following, err := h.followSvc.GetFollowing(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get my following error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToFollowingUserResponses(following)})
}

// GetStats godoc
// @Summary Get follow stats
// @Description Get follower and following counts for a user
// @Tags follows
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} object "Follow stats"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{userId}/follow-stats [get]
func (h *FollowHandler) GetStats(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	followerCount, followingCount, err := h.followSvc.GetStats(c.Request.Context(), int32(userID))
	if err != nil {
		log.Printf("get stats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"follower_count":  followerCount,
			"following_count": followingCount,
		},
	})
}
