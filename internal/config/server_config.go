package config

import (
	"os"
)

type ServerConfig struct {
	Port        string
	DatabaseURL string
	RedisURL    string
}

func LoadServerConfig() *ServerConfig {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		// Default dev URL
		dbUrl = "postgres://postgres:postgres@localhost:5432/colink?sslmode=disable"
	}

	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = "redis://localhost:6379/0"
	}

	return &ServerConfig{
		Port:        port,
		DatabaseURL: dbUrl,
		RedisURL:    redisUrl,
	}
}
