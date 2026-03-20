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
	movieH     *handler.MovieHandler
	ratingH    *handler.RatingHandler
	watchlistH *handler.WatchlistHandler
	commentH   *handler.CommentHandler
	reactionH  *handler.ReactionHandler
	userH      *handler.UserHandler
	jwt        *token.JWTManager
}

func NewServer(db *pgxpool.Pool, authH *handler.AuthHandler, movieH *handler.MovieHandler, ratingH *handler.RatingHandler, watchlistH *handler.WatchlistHandler, commentH *handler.CommentHandler, reactionH *handler.ReactionHandler, userH *handler.UserHandler, jwt *token.JWTManager) *Server {
	s := &Server{
		router:     gin.Default(),
		db:         db,
		authH:      authH,
		movieH:     movieH,
		ratingH:    ratingH,
		watchlistH: watchlistH,
		commentH:   commentH,
		reactionH:  reactionH,
		userH:      userH,
		jwt:        jwt,
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

	// Movie routes (public)
	movies := v1.Group("/movies")
	{
		movies.GET("", s.movieH.GetMovies)
		movies.GET("/:slug", s.movieH.GetMovie)
		movies.GET("/:movieId/ratings", s.ratingH.GetMovieRatings)
		movies.GET("/:movieId/comments", s.commentH.GetMovieComments)
		movies.GET("/:movieId/comments/count", s.commentH.GetMovieCommentCount)
	}

	// Comments (public read)
	comments := v1.Group("/comments")
	{
		comments.GET("/:commentId", s.commentH.GetComment)
		comments.GET("/:commentId/replies", s.commentH.GetReplies)
		comments.GET("/:commentId/reactions", s.reactionH.GetCommentReactions)
		comments.GET("/:commentId/reactions/summary", s.reactionH.GetCommentReactionSummary)
	}

	// Users (public)
	users := v1.Group("/users")
	{
		users.GET("/:userId", s.userH.GetProfile)
		users.GET("/username/:username", s.userH.GetProfileByUsername)
		users.GET("/search", s.userH.SearchUsers)
	}

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware(s.jwt))
	{
		// User profile
		protected.GET("/me", s.authH.GetMe)
		protected.GET("/me/profile", s.userH.GetMyProfile)
		protected.PATCH("/me/profile", s.userH.UpdateMyProfile)

		// Ratings
		protected.GET("/me/ratings", s.ratingH.GetMyRatings)
		protected.GET("/movies/:movieId/rating", s.ratingH.GetMyRating)
		protected.PUT("/movies/:movieId/rating", s.ratingH.RateMovie)
		protected.DELETE("/movies/:movieId/rating", s.ratingH.DeleteRating)

		// Watchlist
		protected.GET("/me/watchlist", s.watchlistH.GetWatchlist)
		protected.GET("/me/watchlist/count", s.watchlistH.GetWatchlistCount)
		protected.POST("/movies/:movieId/watchlist", s.watchlistH.AddToWatchlist)
		protected.GET("/movies/:movieId/watchlist", s.watchlistH.GetWatchlistEntry)
		protected.GET("/movies/:movieId/watchlist/check", s.watchlistH.CheckInWatchlist)
		protected.PATCH("/movies/:movieId/watchlist", s.watchlistH.UpdateWatchlistEntry)
		protected.POST("/movies/:movieId/watchlist/watched", s.watchlistH.MarkAsWatched)
		protected.DELETE("/movies/:movieId/watchlist/watched", s.watchlistH.MarkAsUnwatched)
		protected.DELETE("/movies/:movieId/watchlist", s.watchlistH.RemoveFromWatchlist)

		// Comments
		protected.GET("/me/comments", s.commentH.GetMyComments)
		protected.POST("/movies/:movieId/comments", s.commentH.CreateComment)
		protected.POST("/movies/:movieId/comments/:commentId/replies", s.commentH.CreateReply)
		protected.PATCH("/comments/:commentId", s.commentH.UpdateComment)
		protected.DELETE("/comments/:commentId", s.commentH.DeleteComment)

		// Reactions
		protected.POST("/comments/:commentId/reactions", s.reactionH.ReactToComment)
		protected.DELETE("/comments/:commentId/reactions", s.reactionH.RemoveCommentReaction)
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
