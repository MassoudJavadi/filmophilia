package api

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	router *gin.Engine
	db     *pgxpool.Pool
	userH  *handler.UserHandler
}

// NewServer is a Provider for Wire
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
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", s.userH.Signup)
		}
	}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}