package api

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/handler"
	"github.com/MassoudJavadi/filmophilia/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
    router *gin.Engine
    db     *pgxpool.Pool
    userH  *handler.UserHandler
}

func NewServer(db *pgxpool.Pool, userH *handler.UserHandler) *Server {
    s := &Server{
        router: gin.Default(),
        db:     db,
        userH:  userH,
    }

    s.setupRoutes()
    return s
}

func (s *Server) setupRoutes() {
    v1 := s.router.Group("/api/v1")
    
    // Public routes
    auth := v1.Group("/auth")
    {
        auth.POST("/signup", s.userH.Signup)
        auth.POST("/login", s.userH.Login)
		auth.POST("/refresh", s.userH.Refresh)
    }

    // Protected routes (Protected by JWT)
    protected := v1.Group("/")
    protected.Use(middleware.AuthMiddleware()) // اینجا بادیگارد رو میذاریم
    {
        // مثلاً یه روت تست برای پروفایل
        protected.GET("/me", s.userH.GetMe) 
    }
}

func (s *Server) Start(addr string) error {
    return s.router.Run(addr)
}