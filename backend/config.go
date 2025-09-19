package main

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds DB and JWT settings
type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	JWTSecret  string
}

// LoadConfig loads environment variables from .env (with fallback defaults)
func LoadConfig() Config {
	// Load .env file if present
	_ = godotenv.Load()

	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "yourpassword"), // ⚠️ replace
		DBName:     getEnv("DB_NAME", "filevault"),
		JWTSecret:  getEnv("JWT_SECRET", "supersecretkey"), // ⚠️ replace
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
