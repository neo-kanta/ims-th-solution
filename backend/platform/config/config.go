package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig holds all application configuration loaded from environment variables.
type AppConfig struct {
	Env            string
	Port           string
	LogLevel       string
	JWTSecret      string
	DBHost         string
	DBPort         string
	DBName         string
	DBUser         string
	DBPassword     string
	DBSSLMode      string
	DBMaxConns     int
}

// Load reads configuration from environment variables.
// It attempts to load a .env file if present but does not fail if absent (production uses real env vars).
func Load() (*AppConfig, error) {
	// Best-effort .env loading — ignore error if file doesn't exist
	_ = godotenv.Load()
	maxConns, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_CONNECTIONS", "25"))

	cfg := &AppConfig{
		Env:        getEnvOrDefault("APP_ENV", "development"),
		Port:       getEnvOrDefault("APP_PORT", "8080"),
		LogLevel:   getEnvOrDefault("APP_LOG_LEVEL", "debug"),
		JWTSecret:  os.Getenv("APP_JWT_SECRET"),
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBName:     getEnvOrDefault("DB_NAME", "ims_dev"),
		DBUser:     getEnvOrDefault("DB_USER", "ims_app"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBSSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		DBMaxConns: maxConns,
	}

	if cfg.JWTSecret == "" && cfg.Env != "development" {
		return nil, fmt.Errorf("APP_JWT_SECRET is required in non-development environments")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "dev-secret-change-in-production"
	}

	return cfg, nil
}

// DSN returns the PostgreSQL connection string.
func (c *AppConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
