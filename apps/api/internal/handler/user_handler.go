package handler

import (
	"net/http"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
    userSvc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
    return &UserHandler{userSvc: svc}
}

func (h *UserHandler) Signup(c *gin.Context) {
    var req dto.SignupRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.userSvc.Register(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
        return
    }

    c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Login(c *gin.Context) {
    var req dto.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.userSvc.Login(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, resp)
}


func (h *UserHandler) GetMe(c *gin.Context) {
    // Get data from context set by middleware
    userID := c.MustGet("user_id").(int32)
    role := c.MustGet("user_role").(string)

    c.JSON(http.StatusOK, gin.H{
        "id":   userID,
        "role": role,
        "msg":  "You are authorized!",
    })
}

func (h *UserHandler) Refresh(c *gin.Context) {
    var req dto.RefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.userSvc.RefreshToken(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, resp)
}