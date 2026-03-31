package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// AppConfig holds all application configuration loaded from environment variables.
type AppConfig struct {
	Env        string
	Port       string
	LogLevel   string
	JWTSecret  string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string
	DBMaxConns int

	// IAM: Session policy
	SessionIdleTimeout    time.Duration // Idle timeout for sessions (default: 30m)
	SessionAbsoluteLife   time.Duration // Absolute session lifetime (default: 24h)
	SessionMaxConcurrent  int           // Max concurrent sessions per user (0 = unlimited)
	SessionConcurrentMode string        // "reject" or "evict_oldest" (default: evict_oldest)

	// IAM: MFA
	MFAIssuerName     string // TOTP issuer name shown in authenticator apps
	MFAEncryptionKey  string // AES-256 key for encrypting MFA secrets (hex-encoded, 64 chars)
	MFAForcedForAdmin bool   // Require MFA for IAM_ADMIN users

	// IAM: Rate limiting
	RateLimitLoginPerIP     int           // Max failed logins per IP within window (default: 20)
	RateLimitLoginPerUser   int           // Max failed logins per IP+username within window (default: 5)
	RateLimitLoginWindow    time.Duration // Rate limit window (default: 15m)
	RateLimitCleanupEnabled bool          // Enable periodic cleanup of old attempts

	// IAM: Password policy
	PasswordMaxAgeDays int // Password expiration in days (0 = disabled, default: 0)

	// IAM: JWT key rotation
	JWTSecretPrevious string // Previous JWT secret for rotation window verification
	JWTKeyID          string // Key ID for current JWT secret (default: "k1")
	JWTKeyIDPrevious  string // Key ID of the previous JWT secret (default: "prev")

	// IAM: IP allowlisting
	AdminIPAllowlist []string // CIDR ranges allowed for admin routes (empty = disabled)
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
		DBPort:     getEnvOrDefault("DB_PORT", "5437"),
		DBName:     getEnvOrDefault("DB_NAME", "ims_dev"),
		DBUser:     getEnvOrDefault("DB_USER", "ims_app"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBSSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		DBMaxConns: maxConns,

		// Session policy
		SessionIdleTimeout:    parseDuration("SESSION_IDLE_TIMEOUT", "30m"),
		SessionAbsoluteLife:   parseDuration("SESSION_ABSOLUTE_LIFE", "24h"),
		SessionMaxConcurrent:  parseInt("SESSION_MAX_CONCURRENT", 0),
		SessionConcurrentMode: getEnvOrDefault("SESSION_CONCURRENT_MODE", "evict_oldest"),

		// MFA
		MFAIssuerName:     getEnvOrDefault("MFA_ISSUER_NAME", "IMS Thailand"),
		MFAEncryptionKey:  os.Getenv("MFA_ENCRYPTION_KEY"),
		MFAForcedForAdmin: parseBool("MFA_FORCED_FOR_ADMIN", true),

		// Rate limiting
		RateLimitLoginPerIP:     parseInt("RATE_LIMIT_LOGIN_PER_IP", 20),
		RateLimitLoginPerUser:   parseInt("RATE_LIMIT_LOGIN_PER_USER", 5),
		RateLimitLoginWindow:    parseDuration("RATE_LIMIT_LOGIN_WINDOW", "15m"),
		RateLimitCleanupEnabled: parseBool("RATE_LIMIT_CLEANUP_ENABLED", true),

		// Password policy
		PasswordMaxAgeDays: parseInt("PASSWORD_MAX_AGE_DAYS", 0),

		// JWT key rotation
		JWTSecretPrevious: os.Getenv("APP_JWT_SECRET_PREVIOUS"),
		JWTKeyID:          getEnvOrDefault("APP_JWT_KEY_ID", "k1"),
		JWTKeyIDPrevious:  getEnvOrDefault("APP_JWT_KEY_ID_PREVIOUS", "prev"),

		// IP allowlisting
		AdminIPAllowlist: parseStringSlice("ADMIN_IP_ALLOWLIST"),
	}

	if cfg.JWTSecret == "" && cfg.Env != "development" {
		return nil, fmt.Errorf("APP_JWT_SECRET is required in non-development environments")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "dev-secret-change-in-production"
	}

	// Generate a default MFA encryption key in development only
	if cfg.MFAEncryptionKey == "" && cfg.Env == "development" {
		cfg.MFAEncryptionKey = "0000000000000000000000000000000000000000000000000000000000000000"
	}
	if cfg.MFAEncryptionKey == "" && cfg.Env != "development" {
		return nil, fmt.Errorf("MFA_ENCRYPTION_KEY is required in non-development environments")
	}

	return cfg, nil
}

// PasswordMaxAge returns the password max age as a time.Duration.
func (c *AppConfig) PasswordMaxAge() time.Duration {
	if c.PasswordMaxAgeDays <= 0 {
		return 0
	}
	return time.Duration(c.PasswordMaxAgeDays) * 24 * time.Hour
}

// DSN returns the PostgreSQL connection string (keyword format for pgx).
func (c *AppConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// MigrationDSN returns a URL-style connection string for golang-migrate.
func (c *AppConfig) MigrationDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func parseDuration(key, defaultVal string) time.Duration {
	val := getEnvOrDefault(key, defaultVal)
	d, err := time.ParseDuration(val)
	if err != nil {
		d, _ = time.ParseDuration(defaultVal)
	}
	return d
}

func parseInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

func parseBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

func parseStringSlice(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
