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
	authH  *handler.AuthHandler
}

func NewServer(db *pgxpool.Pool, authH *handler.AuthHandler) *Server {
	s := &Server{
		router: gin.Default(),
		db:     db,
		authH:  authH,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	v1 := s.router.Group("/api/v1")

	// Public routes
	auth := v1.Group("/auth")
	{
		auth.POST("/signup", s.authH.Signup)
		auth.POST("/login", s.authH.Login)
		auth.POST("/refresh", s.authH.Refresh)
	}

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/me", s.authH.GetMe)
	}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}
