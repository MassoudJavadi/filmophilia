//go:build wireinject
// +build wireinject

package api

import (
	"os"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/handler"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/token"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

func provideJWTManager() *token.JWTManager {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}
	return token.NewJWTManager(secret)
}

func InitializeServer(dbPool *pgxpool.Pool) *Server {
	wire.Build(
		wire.Bind(new(db.DBTX), new(*pgxpool.Pool)),
		db.New,
		provideJWTManager,
		service.NewAuthService,
		handler.NewAuthHandler,
		NewServer,
	)
	return &Server{}
}
