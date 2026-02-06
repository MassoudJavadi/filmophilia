package api

import (
	"context"
	"net/http"
	"time"

	"github.com/MassoudJavadi/filmophilia/api/internal/handler"
	"github.com/MassoudJavadi/filmophilia/api/internal/middleware"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/token"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)


type Server struct {
	httpServer *http.Server
	router     *gin.Engine
	db         *pgxpool.Pool
	authH      *handler.AuthHandler
	jwt        *token.JWTManager
}

func NewServer(db *pgxpool.Pool, authH *handler.AuthHandler, jwt *token.JWTManager) *Server {
	s := &Server{
		router: gin.Default(),
		db:     db,
		authH:  authH,
		jwt:    jwt,
	}

	s.router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"}, //Client(Next.js) url
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
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
		auth.POST("/logout", s.authH.Logout)

		auth.GET("/google", s.authH.GoogleRedirect)
        auth.GET("/google/callback", s.authH.GoogleCallback)
	}

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware(s.jwt))
	{
		protected.GET("/me", s.authH.GetMe)
	}
}

func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
