package config

import (
	"log"
	"os"
	"strings"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	JWTSecret   string
	CORSOrigins []string
}

// Load loads configuration from environment variables.
// It panics if required variables are missing in production.
func Load() Config {
	cfg := Config{}

	// JWT Secret - required in production
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		if isProduction() {
			log.Fatal("JWT_SECRET environment variable is required in production")
		}
		log.Println("WARNING: JWT_SECRET not set, using insecure default for development")
		cfg.JWTSecret = generateDevSecret()
	}

	// CORS Origins - comma-separated list
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		if isProduction() {
			log.Fatal("CORS_ORIGINS environment variable is required in production")
		}
		cfg.CORSOrigins = []string{"http://localhost:3000"}
	} else {
		cfg.CORSOrigins = parseOrigins(corsOrigins)
	}

	return cfg
}

func isProduction() bool {
	env := os.Getenv("GIN_MODE")
	return env == "release"
}

func parseOrigins(origins string) []string {
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func generateDevSecret() string {
	// Use a hash of hostname + fixed string to avoid identical secrets across machines
	// while still being deterministic for dev convenience
	hostname, _ := os.Hostname()
	return "dev-" + hostname + "-insecure-secret"
}
