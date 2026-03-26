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

type UserHandler struct {
	userSvc service.UserService
}

func NewUserHandler(userSvc service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get a user's public profile by ID
// @Tags users
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.UserProfileResponse} "User profile"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid user ID"
// @Failure 404 {object} dto.NotFoundErrorResponse "User not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /users/{userId} [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "invalid user id")
		return
	}

	profile, err := h.userSvc.GetProfile(c.Request.Context(), int32(userID))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
			return
		}
		log.Printf("get profile error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToUserProfileResponse(profile))
}

// GetProfileByUsername godoc
// @Summary Get user profile by username
// @Description Get a user's public profile by username
// @Tags users
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} dto.SuccessResponse{data=dto.UserProfileResponse} "User profile"
// @Failure 400 {object} dto.BadRequestErrorResponse "Username required"
// @Failure 404 {object} dto.NotFoundErrorResponse "User not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /users/username/{username} [get]
func (h *UserHandler) GetProfileByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.Error(c, http.StatusBadRequest, "USERNAME_REQUIRED", "username is required")
		return
	}

	profile, err := h.userSvc.GetProfileByUsername(c.Request.Context(), username)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
			return
		}
		log.Printf("get profile by username error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToUserProfileResponseFromUsername(profile))
}

// GetMyProfile godoc
// @Summary Get my profile
// @Description Get the authenticated user's profile
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.UserProfileResponse} "User profile"
// @Failure 404 {object} dto.NotFoundErrorResponse "User not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /me/profile [get]
func (h *UserHandler) GetMyProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	profile, err := h.userSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
			return
		}
		log.Printf("get my profile error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToUserProfileResponse(profile))
}

// UpdateMyProfile godoc
// @Summary Update my profile
// @Description Update the authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "Profile updates"
// @Success 200 {object} dto.SuccessResponse{data=dto.UserResponse} "Updated user"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid request"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /me/profile [patch]
func (h *UserHandler) UpdateMyProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	user, err := h.userSvc.UpdateProfile(c.Request.Context(), userID, req.DisplayName, req.AvatarURL, req.Bio)
	if err != nil {
		log.Printf("update profile error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToUserResponse(user))
}

// SearchUsers godoc
// @Summary Search users
// @Description Search for users by username or display name
// @Tags users
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.UserSearchResult} "Search results"
// @Failure 400 {object} dto.BadRequestErrorResponse "Search query required"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "SEARCH_QUERY_REQUIRED", "search query is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	users, err := h.userSvc.SearchUsers(c.Request.Context(), query, int32(limit), int32(offset))
	if err != nil {
		log.Printf("search users error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToUserSearchResults(users))
}
