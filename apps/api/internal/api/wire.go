//go:build wireinject
// +build wireinject

package api

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/handler"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitializeServer(dbPool *pgxpool.Pool) *Server {
	wire.Build(
		wire.Bind(new(db.DBTX), new(*pgxpool.Pool)), // Bind pgxpool.Pool to DBTX interface
		db.New,                 // SQLC Queries
		service.NewUserService, // Service layer
		handler.NewUserHandler, // Handler layer
		NewServer,              // Server engine
	)
	return &Server{}
}