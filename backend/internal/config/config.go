package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort     string
	DatabaseURL    string
	JWTSecret      string
	JWTExpiration  string
	UploadDir      string
	AllowedOrigins string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/expensesplit?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
		JWTExpiration:  getEnv("JWT_EXPIRATION", "24h"),
		UploadDir:      getEnv("UPLOAD_DIR", "./uploads"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
