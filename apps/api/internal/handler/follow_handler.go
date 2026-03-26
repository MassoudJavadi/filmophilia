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
// @Success 201 {object} dto.SuccessResponse{data=dto.MessageData} "Followed successfully"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid user ID or cannot follow self"
// @Failure 404 {object} dto.NotFoundErrorResponse "User not found"
// @Failure 409 {object} dto.ConflictErrorResponse "Already following"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /users/{userId}/follow [post]
func (h *FollowHandler) Follow(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	targetIDStr := c.Param("userId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "invalid user id")
		return
	}

	if err := h.followSvc.Follow(c.Request.Context(), userID, int32(targetID)); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotFollowSelf):
			response.Error(c, http.StatusBadRequest, "CANNOT_FOLLOW_SELF", "cannot follow yourself")
		case errors.Is(err, service.ErrAlreadyFollowing):
			response.Error(c, http.StatusConflict, "ALREADY_FOLLOWING", "already following this user")
		case errors.Is(err, service.ErrUserNotFound):
			response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		default:
			log.Printf("follow error: %v", err)
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		}
		return
	}

	response.Message(c, http.StatusCreated, "followed successfully")
}

// Unfollow godoc
// @Summary Unfollow a user
// @Description Stop following another user
// @Tags follows
// @Produce json
// @Security BearerAuth
// @Param userId path int true "User ID to unfollow"
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageData} "Unfollowed successfully"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid user ID or cannot unfollow self"
// @Failure 404 {object} dto.NotFoundErrorResponse "Not following this user"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /users/{userId}/follow [delete]
func (h *FollowHandler) Unfollow(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	targetIDStr := c.Param("userId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "invalid user id")
		return
	}

	if err := h.followSvc.Unfollow(c.Request.Context(), userID, int32(targetID)); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotFollowSelf):
			response.Error(c, http.StatusBadRequest, "CANNOT_UNFOLLOW_SELF", "cannot unfollow yourself")
		case errors.Is(err, service.ErrNotFollowing):
			response.Error(c, http.StatusNotFound, "NOT_FOLLOWING", "not following this user")
		default:
			log.Printf("unfollow error: %v", err)
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		}
		return
	}

	response.Message(c, http.StatusOK, "unfollowed successfully")
}

// IsFollowing godoc
// @Summary Check if following
// @Description Check if the authenticated user is following another user
// @Tags follows
// @Produce json
// @Security BearerAuth
// @Param userId path int true "User ID to check"
// @Success 200 {object} dto.SuccessResponse{data=dto.FollowStatusData} "Following status"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid user ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /users/{userId}/following/check [get]
func (h *FollowHandler) IsFollowing(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	targetIDStr := c.Param("userId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "invalid user id")
		return
	}

	isFollowing, err := h.followSvc.IsFollowing(c.Request.Context(), userID, int32(targetID))
	if err != nil {
		log.Printf("is following error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, dto.FollowStatusData{IsFollowing: isFollowing})
}

// GetFollowers godoc
// @Summary Get user's followers
// @Description Get list of users following a specific user
// @Tags follows
// @Produce json
// @Param userId path int true "User ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.FollowUserResponse} "List of followers"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid user ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /users/{userId}/followers [get]
func (h *FollowHandler) GetFollowers(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "invalid user id")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	followers, err := h.followSvc.GetFollowers(c.Request.Context(), int32(userID), int32(limit), int32(offset))
	if err != nil {
		log.Printf("get followers error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToFollowUserResponses(followers))
}

// GetFollowing godoc
// @Summary Get user's following
// @Description Get list of users that a specific user is following
// @Tags follows
// @Produce json
// @Param userId path int true "User ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.FollowUserResponse} "List of following"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid user ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /users/{userId}/following [get]
func (h *FollowHandler) GetFollowing(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "invalid user id")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	following, err := h.followSvc.GetFollowing(c.Request.Context(), int32(userID), int32(limit), int32(offset))
	if err != nil {
		log.Printf("get following error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToFollowingUserResponses(following))
}

// GetMyFollowers godoc
// @Summary Get my followers
// @Description Get list of users following the authenticated user
// @Tags follows
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.FollowUserResponse} "List of followers"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /me/followers [get]
func (h *FollowHandler) GetMyFollowers(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	followers, err := h.followSvc.GetFollowers(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get my followers error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToFollowUserResponses(followers))
}

// GetMyFollowing godoc
// @Summary Get my following
// @Description Get list of users the authenticated user is following
// @Tags follows
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.FollowUserResponse} "List of following"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /me/following [get]
func (h *FollowHandler) GetMyFollowing(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	following, err := h.followSvc.GetFollowing(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get my following error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToFollowingUserResponses(following))
}

// GetStats godoc
// @Summary Get follow stats
// @Description Get follower and following counts for a user
// @Tags follows
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.FollowStatsResponse} "Follow stats"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid user ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /users/{userId}/follow-stats [get]
func (h *FollowHandler) GetStats(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "invalid user id")
		return
	}

	followerCount, followingCount, err := h.followSvc.GetStats(c.Request.Context(), int32(userID))
	if err != nil {
		log.Printf("get stats error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, dto.FollowStatsResponse{
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
	})
}
