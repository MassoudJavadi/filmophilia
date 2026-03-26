package handler

import (
	"crypto/subtle"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/oauth"
	"github.com/MassoudJavadi/filmophilia/api/internal/response"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc  *service.AuthService
	oauthSvc *service.OAuthService
}

func NewAuthHandler(as *service.AuthService, os *service.OAuthService) *AuthHandler {
	return &AuthHandler{authSvc: as, oauthSvc: os}
}

// Signup godoc
// @Summary Register a new user
// @Description Create a new user account with email, username, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.SignupRequest true "Signup credentials"
// @Success 201 {object} dto.SuccessResponse{data=dto.UserResponse} "User created"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid request body"
// @Failure 409 {object} dto.ConflictErrorResponse "Email or username already exists"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /auth/signup [post]
func (h *AuthHandler) Signup(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	user, err := h.authSvc.Signup(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailExists):
			response.Error(c, http.StatusConflict, "EMAIL_ALREADY_REGISTERED", "email already registered")
		case errors.Is(err, service.ErrUsernameExists):
			response.Error(c, http.StatusConflict, "USERNAME_ALREADY_TAKEN", "username already taken")
		default:
			log.Printf("signup error: %v", err)
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		}
		return
	}
	response.Created(c, mapper.ToUserResponse(user))
}

// Login godoc
// @Summary Authenticate user
// @Description Login with email and password to get access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.SuccessResponse{data=dto.AuthResponse} "Authenticated successfully"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid request body"
// @Failure 401 {object} dto.UnauthorizedErrorResponse "Invalid credentials or user banned"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	resp, err := h.authSvc.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrUserBanned) {
			response.Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
			return
		}
		log.Printf("login error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}
	response.OK(c, resp)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Get a new access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh token"
// @Success 200 {object} dto.SuccessResponse{data=dto.AuthResponse} "Token refreshed"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid request body"
// @Failure 401 {object} dto.UnauthorizedErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	resp, err := h.authSvc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			response.Error(c, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", err.Error())
			return
		}
		log.Printf("refresh error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}
	response.OK(c, resp)
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate the refresh token to logout the user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh token to invalidate"
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageData} "Logged out successfully"
// @Failure 400 {object} dto.BadRequestErrorResponse "Refresh token required"
// @Failure 500 {object} dto.InternalServerErrorResponse "Failed to logout"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshRequest //Get refresh token from body to invalidate it.
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "REFRESH_TOKEN_REQUIRED", "refresh token required")
		return
	}

	if err := h.authSvc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, http.StatusInternalServerError, "LOGOUT_FAILED", "failed to logout")
		return
	}

	response.Message(c, http.StatusOK, "logged out successfully")
}

// GetMe godoc
// @Summary Get current user
// @Description Get the authenticated user's information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.UserResponse} "Current user"
// @Failure 404 {object} dto.NotFoundErrorResponse "User not found"
// @Router /me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	user, err := h.authSvc.GetUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}

	response.OK(c, mapper.ToUserResponse(user))
}

func (h *AuthHandler) GoogleRedirect(c *gin.Context) {
	state, err := oauth.GenerateState(32)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "STATE_GENERATION_FAILED", "failed to generate state")
		return
	}

	// Store state in cookie for 15 minutes
	// Secure flag enabled in production (GIN_MODE=release)
	secure := os.Getenv("GIN_MODE") == "release"
	c.SetCookie("oauth_state", state, 900, "/", "", secure, true)

	url := h.oauthSvc.GetGoogleAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	cookieState, err := c.Cookie("oauth_state")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "STATE_COOKIE_NOT_FOUND", "state cookie not found")
		return
	}

	queryState := c.Query("state")
	if queryState == "" {
		response.Error(c, http.StatusBadRequest, "STATE_PARAMETER_MISSING", "state parameter missing")
		return
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(cookieState), []byte(queryState)) != 1 {
		response.Error(c, http.StatusForbidden, "INVALID_OAUTH_STATE", "invalid oauth state")
		return
	}

	// Remove cookie after use (secure flag matches redirect)
	secure := os.Getenv("GIN_MODE") == "release"
	c.SetCookie("oauth_state", "", -1, "/", "", secure, true)

	code := c.Query("code")
	if code == "" {
		response.Error(c, http.StatusBadRequest, "AUTHORIZATION_CODE_MISSING", "authorization code missing")
		return
	}

	resp, err := h.oauthSvc.HandleGoogleCallback(c.Request.Context(), code)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GOOGLE_CALLBACK_FAILED", err.Error())
		return
	}

	response.OK(c, resp)
}
