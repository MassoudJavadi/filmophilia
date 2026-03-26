//go:build wireinject
// +build wireinject

package api

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/handler"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/config"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/oauth"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/ratelimit"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/token"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

func provideConfig() config.Config {
	return config.Load()
}

func provideJWTManager(cfg config.Config) *token.JWTManager {
	return token.NewJWTManager(cfg.JWTSecret)
}

func provideRateLimiter() *ratelimit.RateLimiter {
	client := ratelimit.NewRedisClient()
	return ratelimit.NewRateLimiter(client)
}

func provideRateLimitConfig() ratelimit.Config {
	return ratelimit.LoadFromEnv()
}

func InitializeServer(dbPool *pgxpool.Pool) *Server {
	wire.Build(
		wire.Bind(new(db.DBTX), new(*pgxpool.Pool)),
		db.New,
		provideConfig,
		provideJWTManager,
		provideRateLimiter,
		provideRateLimitConfig,
		oauth.NewGoogleManager,
		service.NewAuthService,
		service.NewOAuthService,
		service.NewMovieService,
		service.NewRatingService,
		service.NewWatchlistService,
		service.NewCommentService,
		service.NewReactionService,
		service.NewUserService,
		service.NewFollowService,
		service.NewGenreService,
		service.NewActivityService,
		service.NewNotificationService,
		handler.NewAuthHandler,
		handler.NewMovieHandler,
		handler.NewRatingHandler,
		handler.NewWatchlistHandler,
		handler.NewCommentHandler,
		handler.NewReactionHandler,
		handler.NewUserHandler,
		handler.NewFollowHandler,
		handler.NewGenreHandler,
		handler.NewActivityHandler,
		handler.NewNotificationHandler,
		NewServer,
	)
	return &Server{}
}
