package handler

import (
	"net/http"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(us *service.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

func (h *UserHandler) Signup(c *gin.Context) {
	var req dto.SignupRequest

	// 1. Bind JSON and Validate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Call Service
	user, err := h.userService.Register(c.Request.Context(), req)
	if err != nil {
	
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	// 3. Return Success (Using a Map or DTO to hide password)
	c.JSON(http.StatusCreated, dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
	})
}