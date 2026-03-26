package api

import (
	"context"
	"net/http"
	"time"

	"github.com/MassoudJavadi/filmophilia/api/internal/handler"
	"github.com/MassoudJavadi/filmophilia/api/internal/middleware"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/config"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/ratelimit"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/token"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	httpServer  *http.Server
	router      *gin.Engine
	db          *pgxpool.Pool
	config      config.Config
	authH       *handler.AuthHandler
	movieH      *handler.MovieHandler
	ratingH     *handler.RatingHandler
	watchlistH  *handler.WatchlistHandler
	commentH    *handler.CommentHandler
	reactionH   *handler.ReactionHandler
	userH       *handler.UserHandler
	followH     *handler.FollowHandler
	genreH      *handler.GenreHandler
	activityH   *handler.ActivityHandler
	notifH      *handler.NotificationHandler
	jwt         *token.JWTManager
	rateLimiter *ratelimit.RateLimiter
	rlConfig    ratelimit.Config
}

func NewServer(db *pgxpool.Pool, cfg config.Config, authH *handler.AuthHandler, movieH *handler.MovieHandler, ratingH *handler.RatingHandler, watchlistH *handler.WatchlistHandler, commentH *handler.CommentHandler, reactionH *handler.ReactionHandler, userH *handler.UserHandler, followH *handler.FollowHandler, genreH *handler.GenreHandler, activityH *handler.ActivityHandler, notifH *handler.NotificationHandler, jwt *token.JWTManager, rateLimiter *ratelimit.RateLimiter, rlConfig ratelimit.Config) *Server {
	s := &Server{
		router:      gin.Default(),
		db:          db,
		config:      cfg,
		authH:       authH,
		movieH:      movieH,
		ratingH:     ratingH,
		watchlistH:  watchlistH,
		commentH:    commentH,
		reactionH:   reactionH,
		userH:       userH,
		followH:     followH,
		genreH:      genreH,
		activityH:   activityH,
		notifH:      notifH,
		jwt:         jwt,
		rateLimiter: rateLimiter,
		rlConfig:    rlConfig,
	}

	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	s.router.Use(middleware.SecurityHeadersMiddleware())
	s.router.Use(middleware.PrometheusMiddleware())

	// Prometheus metrics endpoint
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation endpoint
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	v1 := s.router.Group("/api/v1")

	// Rate limit middleware helpers
	rl := s.rateLimiter
	cfg := s.rlConfig
	window := ratelimit.HourWindow

	ipRL := func(limit int) gin.HandlerFunc {
		return middleware.RateLimitMiddleware(rl, limit, window, ratelimit.IPKey("ip"))
	}
	userRL := func(limit int) gin.HandlerFunc {
		return middleware.RateLimitMiddleware(rl, limit, window, ratelimit.UserKey("user"))
	}

	// Public routes
	auth := v1.Group("/auth")
	{
		auth.POST("/signup", ipRL(cfg.SignupLimit), s.authH.Signup)
		auth.POST("/login", ipRL(cfg.LoginLimit), s.authH.Login)
		auth.POST("/refresh", ipRL(cfg.PublicReadLimit), s.authH.Refresh)
		auth.POST("/logout", ipRL(cfg.PublicReadLimit), s.authH.Logout)

		auth.GET("/google", ipRL(cfg.PublicReadLimit), s.authH.GoogleRedirect)
		auth.GET("/google/callback", ipRL(cfg.PublicReadLimit), s.authH.GoogleCallback)
	}

	// Movie routes (public)
	movies := v1.Group("/movies")
	movies.Use(ipRL(cfg.PublicReadLimit))
	{
		movies.GET("", s.movieH.GetMovies)
		movies.GET("/:slug", s.movieH.GetMovie)
		movies.GET("/:movieId/ratings", s.ratingH.GetMovieRatings)
		movies.GET("/:movieId/comments", s.commentH.GetMovieComments)
		movies.GET("/:movieId/comments/count", s.commentH.GetMovieCommentCount)
	}
	// Search has its own stricter limit
	v1.GET("/movies/search", ipRL(cfg.SearchLimit), s.movieH.AdvancedSearch)

	// Comments (public read)
	comments := v1.Group("/comments")
	comments.Use(ipRL(cfg.PublicReadLimit))
	{
		comments.GET("/:commentId", s.commentH.GetComment)
		comments.GET("/:commentId/replies", s.commentH.GetReplies)
		comments.GET("/:commentId/reactions", s.reactionH.GetCommentReactions)
		comments.GET("/:commentId/reactions/summary", s.reactionH.GetCommentReactionSummary)
	}

	// Users (public)
	users := v1.Group("/users")
	users.Use(ipRL(cfg.PublicReadLimit))
	{
		users.GET("/:userId", s.userH.GetProfile)
		users.GET("/username/:username", s.userH.GetProfileByUsername)
		users.GET("/:userId/followers", s.followH.GetFollowers)
		users.GET("/:userId/following", s.followH.GetFollowing)
		users.GET("/:userId/follow-stats", s.followH.GetStats)
		users.GET("/:userId/activities", s.activityH.GetUserActivities)
	}
	// User search has its own stricter limit
	v1.GET("/users/search", ipRL(cfg.SearchLimit), s.userH.SearchUsers)

	// Movie activities (public)
	v1.GET("/movies/:movieId/activities", ipRL(cfg.PublicReadLimit), s.activityH.GetMovieActivities)

	// Genres (public)
	genres := v1.Group("/genres")
	genres.Use(ipRL(cfg.PublicReadLimit))
	{
		genres.GET("", s.genreH.GetAll)
		genres.GET("/:genreId", s.genreH.GetByID)
		genres.GET("/:genreId/movies", s.genreH.GetMoviesByGenre)
		genres.GET("/:genreId/stats", s.genreH.GetGenreWithCount)
		genres.GET("/slug/:slug", s.genreH.GetBySlug)
		genres.GET("/slug/:slug/movies", s.genreH.GetMoviesByGenreSlug)
	}

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware(s.jwt))
	{
		// User profile (protected reads)
		protected.GET("/me", userRL(cfg.ProtectedRead), s.authH.GetMe)
		protected.GET("/me/profile", userRL(cfg.ProtectedRead), s.userH.GetMyProfile)
		protected.PATCH("/me/profile", userRL(cfg.ProtectedWrite), s.userH.UpdateMyProfile)

		// Ratings (protected)
		protected.GET("/me/ratings", userRL(cfg.ProtectedRead), s.ratingH.GetMyRatings)
		protected.GET("/movies/:movieId/rating", userRL(cfg.ProtectedRead), s.ratingH.GetMyRating)
		protected.PUT("/movies/:movieId/rating", userRL(cfg.ProtectedWrite), s.ratingH.RateMovie)
		protected.DELETE("/movies/:movieId/rating", userRL(cfg.ProtectedWrite), s.ratingH.DeleteRating)

		// Watchlist (protected)
		protected.GET("/me/watchlist", userRL(cfg.ProtectedRead), s.watchlistH.GetWatchlist)
		protected.GET("/me/watchlist/count", userRL(cfg.ProtectedRead), s.watchlistH.GetWatchlistCount)
		protected.POST("/movies/:movieId/watchlist", userRL(cfg.ProtectedWrite), s.watchlistH.AddToWatchlist)
		protected.GET("/movies/:movieId/watchlist", userRL(cfg.ProtectedRead), s.watchlistH.GetWatchlistEntry)
		protected.GET("/movies/:movieId/watchlist/check", userRL(cfg.ProtectedRead), s.watchlistH.CheckInWatchlist)
		protected.PATCH("/movies/:movieId/watchlist", userRL(cfg.ProtectedWrite), s.watchlistH.UpdateWatchlistEntry)
		protected.POST("/movies/:movieId/watchlist/watched", userRL(cfg.ProtectedWrite), s.watchlistH.MarkAsWatched)
		protected.DELETE("/movies/:movieId/watchlist/watched", userRL(cfg.ProtectedWrite), s.watchlistH.MarkAsUnwatched)
		protected.DELETE("/movies/:movieId/watchlist", userRL(cfg.ProtectedWrite), s.watchlistH.RemoveFromWatchlist)

		// Comments (protected)
		protected.GET("/me/comments", userRL(cfg.ProtectedRead), s.commentH.GetMyComments)
		protected.POST("/movies/:movieId/comments", userRL(cfg.ProtectedWrite), s.commentH.CreateComment)
		// Comment replies have stricter limit
		protected.POST("/movies/:movieId/comments/:commentId/replies", userRL(cfg.CommentReplyLimit), s.commentH.CreateReply)
		protected.PATCH("/comments/:commentId", userRL(cfg.ProtectedWrite), s.commentH.UpdateComment)
		protected.DELETE("/comments/:commentId", userRL(cfg.ProtectedWrite), s.commentH.DeleteComment)

		// Reactions (protected writes)
		protected.POST("/comments/:commentId/reactions", userRL(cfg.ProtectedWrite), s.reactionH.ReactToComment)
		protected.DELETE("/comments/:commentId/reactions", userRL(cfg.ProtectedWrite), s.reactionH.RemoveCommentReaction)

		// Follows (protected)
		protected.GET("/me/followers", userRL(cfg.ProtectedRead), s.followH.GetMyFollowers)
		protected.GET("/me/following", userRL(cfg.ProtectedRead), s.followH.GetMyFollowing)
		protected.POST("/users/:userId/follow", userRL(cfg.ProtectedWrite), s.followH.Follow)
		protected.DELETE("/users/:userId/follow", userRL(cfg.ProtectedWrite), s.followH.Unfollow)
		protected.GET("/users/:userId/following/check", userRL(cfg.ProtectedRead), s.followH.IsFollowing)

		// Activities (protected)
		protected.GET("/me/activities", userRL(cfg.ProtectedRead), s.activityH.GetMyActivities)
		protected.GET("/me/feed", userRL(cfg.ProtectedRead), s.activityH.GetFollowingFeed)
		protected.POST("/activities", userRL(cfg.ProtectedWrite), s.activityH.CreateActivity)

		// Notifications (protected)
		protected.GET("/me/notifications", userRL(cfg.ProtectedRead), s.notifH.GetMyNotifications)
		protected.GET("/me/notifications/unread", userRL(cfg.ProtectedRead), s.notifH.GetUnreadNotifications)
		protected.GET("/me/notifications/unread/count", userRL(cfg.ProtectedRead), s.notifH.GetUnreadCount)
		protected.PATCH("/notifications/:notificationId/read", userRL(cfg.ProtectedWrite), s.notifH.MarkAsRead)
		protected.POST("/notifications/read-all", userRL(cfg.ProtectedWrite), s.notifH.MarkAllAsRead)
		protected.DELETE("/notifications/:notificationId", userRL(cfg.ProtectedWrite), s.notifH.DeleteNotification)
		protected.DELETE("/notifications", userRL(cfg.ProtectedWrite), s.notifH.DeleteAllNotifications)
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
