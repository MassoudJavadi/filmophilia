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
        // Here you can check for specific errors like ErrUserBanned
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, resp)
}