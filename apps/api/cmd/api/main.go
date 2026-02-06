package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/MassoudJavadi/filmophilia/api/internal/api"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found, relying on system environment variables")
	}

	// 2. Setup Database Connection with Context
	// We use a timeout context to ensure we don't wait forever for the DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set in environment")
	}

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Ping the database to ensure connection is live
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}
	fmt.Println("✅ Successfully connected to PostgreSQL (Filmophilia DB)")

	// 3. Initialize Server using Google Wire
	// This magically wires up all dependencies: DB -> Queries -> Services -> Handlers -> Server
	server := api.InitializeServer(dbPool)

	// 4. Start the Engine
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("🚀 Filmophilia API starting on port %s\n", port)
	if err := server.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}