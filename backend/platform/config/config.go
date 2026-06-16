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

	// Rate limiting: Global
	RateLimitGlobalPerIP  int           // Max requests per IP across all endpoints (default: 200)
	RateLimitGlobalWindow time.Duration // Global rate limit window (default: 1m)

	// Rate limiting: Login
	RateLimitLoginPerIP   int           // Max login attempts per IP within window (default: 20)
	RateLimitLoginPerUser int           // Max login attempts per IP+username within window (default: 5)
	RateLimitLoginWindow  time.Duration // Login rate limit window (default: 15m)

	// Rate limiting: Refresh
	RateLimitRefreshPerIP  int           // Max refresh attempts per IP (default: 30)
	RateLimitRefreshWindow time.Duration // Refresh rate limit window (default: 15m)

	// Rate limiting: Sensitive (change-password, MFA verify/disable)
	RateLimitSensitiveMax    int           // Max sensitive ops per user (default: 5)
	RateLimitSensitiveWindow time.Duration // Sensitive rate limit window (default: 15m)

	// Rate limiting: Admin
	RateLimitAdminMax    int           // Max admin requests per user (default: 60)
	RateLimitAdminWindow time.Duration // Admin rate limit window (default: 1m)

	// Rate limiting: Export
	RateLimitExportMax    int           // Max export requests per user (default: 5)
	RateLimitExportWindow time.Duration // Export rate limit window (default: 1h)

	// Rate limiting: Infrastructure
	RateLimitBackend string // "memory" or "redis" (default: memory)

	// Redis
	RedisAddr         string        // Redis address (default: localhost:6379)
	RedisPassword     string        // Redis password (default: empty)
	RedisDB           int           // Redis database number (default: 0)
	RedisTLSEnabled   bool          // Enable TLS for Redis connection (default: false)
	RedisDialTimeout  time.Duration // Dial timeout (default: 5s)
	RedisReadTimeout  time.Duration // Read timeout (default: 3s)
	RedisWriteTimeout time.Duration // Write timeout (default: 3s)
	RedisPoolSize     int           // Connection pool size (default: 10)

	// IAM: Account lockout
	LoginMaxFailedAttempts int           // Failed attempts before lockout (default: 10)
	LoginLockoutDuration   time.Duration // How long account stays locked (default: 30m)

	// IAM: Trusted proxies
	TrustedProxies []string // CIDR ranges of trusted reverse proxies

	// IAM: Password policy
	PasswordMaxAgeDays int // Password expiration in days (0 = disabled, default: 0)

	// IAM: JWT key rotation
	JWTSecretPrevious string // Previous JWT secret for rotation window verification
	JWTKeyID          string // Key ID for current JWT secret (default: "k1")
	JWTKeyIDPrevious  string // Key ID of the previous JWT secret (default: "prev")

	// IAM: IP allowlisting
	AdminIPAllowlist []string // CIDR ranges allowed for admin routes (empty = disabled)

	// CORS
	CORSAllowedOrigins []string // Allowed CORS origins (empty = localhost dev defaults)

	// Market data: providers + transport
	MarketDataPrimaryProvider  string        // primary provider tag (e.g., alpha_vantage)
	MarketDataFallbackProvider string        // optional fallback provider tag
	MarketDataHTTPTimeout      time.Duration // per-call HTTP timeout for provider calls
	MarketDataCacheTTL         time.Duration // in-memory cache TTL for repeated quotes

	// Intraday market data on the investment Holdings page.
	// MarketDataStaleAfter is the maximum age of a provider quote before the
	// dashboard surfaces it as stale (defaults to 15m).
	// MarketDataRefreshInterval is reserved for a future scheduler — the
	// service reads it via the IntradayConfig and exposes it on the status
	// endpoint so the UI can hint at how often to poll.
	MarketDataStaleAfter      time.Duration
	MarketDataRefreshInterval time.Duration

	// Optional Twelve Data API key. Twelve Data is documented as a future
	// provider option; this field is captured so the env example matches the
	// spec, but no client is wired yet.
	TwelveDataAPIKey string

	// Market data: Alpha Vantage
	AlphaVantageAPIKey          string        // Alpha Vantage API key. Never logged. REQUIRED in non-development.
	AlphaVantageRateLimitPerMin int           // Per-minute provider call budget (default: 5 — free tier).
	AlphaVantageRateLimitPerDay int           // Per-day provider call budget (default: 500 — free tier).
	AlphaVantageBaseURL         string        // Base URL for the provider (default: https://www.alphavantage.co).
	AlphaVantageTimeout         time.Duration // Per-call timeout (default: 30s).

	// Market data: ingestion scheduler
	IngestionScheduleCron string // Cron expression for the ingestion scheduler (default: "0 18 * * 1-5" — 18:00 weekdays).

	// Investment: valuation
	// ValuationStaleThresholdsByAssetClass maps an asset class code (e.g.
	// "EQUITY", "FIXED_INCOME") to the maximum age of its inputs (price /
	// FX) before the runner records investment_valuation_stale_inputs_total
	// and surfaces the snapshot as stale. Codes match
	// investment__asset_classes; an unknown code falls back to default 1d.
	ValuationStaleThresholdsByAssetClass map[string]time.Duration

	// Chat: provider selection + credentials.
	//
	// LLMProvider chooses which adapter the chat module uses. Slice A only
	// supports "anthropic"; Slice B adds openai, gemini, openai_compatible.
	// Switching at runtime requires only LLM_PROVIDER + the matching
	// *_API_KEY — no code change.
	//
	// AnthropicAPIKey is required when LLMProvider=anthropic. If missing,
	// the chat module logs a warning and is not mounted; the rest of the
	// API continues to function.
	LLMProvider          string
	AnthropicAPIKey      string
	AnthropicModel       string
	ChatMaxTokensPerTurn int

	// Chat: MCP + tool execution.
	//
	// ChatWriteEnabled gates mutating MCP tools. OFF by default — the
	// assistant is read-only unless explicitly enabled AND the user passes a
	// server-side permission check.
	//
	// ChatMCPServersConfig is an optional path to mcp-servers.yaml. When
	// empty, the chat module spawns the bundled ims-mcp stdio server.
	//
	// ChatMCPIMSBin is the path to the ims-mcp binary (default: alongside the
	// server binary). IMSAPIBaseURL is forwarded to it so its read-only tools
	// reach the IMS REST API.
	ChatWriteEnabled     bool
	ChatMCPServersConfig string
	ChatMCPIMSBin        string
	IMSAPIBaseURL        string
}

// Load reads configuration from environment variables.
func Load() (*AppConfig, error) {
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
		MFAForcedForAdmin: parseBool("MFA_FORCED_FOR_ADMIN", false),

		// Rate limiting: Global
		RateLimitGlobalPerIP:  parseInt("RATE_LIMIT_GLOBAL_PER_IP", 200),
		RateLimitGlobalWindow: parseDuration("RATE_LIMIT_GLOBAL_WINDOW", "1m"),

		// Rate limiting: Login
		RateLimitLoginPerIP:   parseInt("RATE_LIMIT_LOGIN_PER_IP", 20),
		RateLimitLoginPerUser: parseInt("RATE_LIMIT_LOGIN_PER_USER", 5),
		RateLimitLoginWindow:  parseDuration("RATE_LIMIT_LOGIN_WINDOW", "1m"),

		// Rate limiting: Refresh
		RateLimitRefreshPerIP:  parseInt("RATE_LIMIT_REFRESH_PER_IP", 30),
		RateLimitRefreshWindow: parseDuration("RATE_LIMIT_REFRESH_WINDOW", "15m"),

		// Rate limiting: Sensitive
		RateLimitSensitiveMax:    parseInt("RATE_LIMIT_SENSITIVE_MAX", 5),
		RateLimitSensitiveWindow: parseDuration("RATE_LIMIT_SENSITIVE_WINDOW", "15m"),

		// Rate limiting: Admin
		RateLimitAdminMax:    parseInt("RATE_LIMIT_ADMIN_MAX", 60),
		RateLimitAdminWindow: parseDuration("RATE_LIMIT_ADMIN_WINDOW", "1m"),

		// Rate limiting: Export
		RateLimitExportMax:    parseInt("RATE_LIMIT_EXPORT_MAX", 5),
		RateLimitExportWindow: parseDuration("RATE_LIMIT_EXPORT_WINDOW", "1h"),

		// Rate limiting: Infrastructure
		RateLimitBackend: getEnvOrDefault("RATE_LIMIT_BACKEND", "memory"),

		// Redis
		RedisAddr:         getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		RedisDB:           parseInt("REDIS_DB", 0),
		RedisTLSEnabled:   parseBool("REDIS_TLS_ENABLED", false),
		RedisDialTimeout:  parseDuration("REDIS_DIAL_TIMEOUT", "5s"),
		RedisReadTimeout:  parseDuration("REDIS_READ_TIMEOUT", "3s"),
		RedisWriteTimeout: parseDuration("REDIS_WRITE_TIMEOUT", "3s"),
		RedisPoolSize:     parseInt("REDIS_POOL_SIZE", 10),

		// Account lockout
		LoginMaxFailedAttempts: parseInt("LOGIN_MAX_FAILED_ATTEMPTS", 10),
		LoginLockoutDuration:   parseDuration("LOGIN_LOCKOUT_DURATION", "30m"),

		// Trusted proxies
		TrustedProxies: parseStringSlice("TRUSTED_PROXIES"),

		// Password policy
		PasswordMaxAgeDays: parseInt("PASSWORD_MAX_AGE_DAYS", 0),

		// JWT key rotation
		JWTSecretPrevious: os.Getenv("APP_JWT_SECRET_PREVIOUS"),
		JWTKeyID:          getEnvOrDefault("APP_JWT_KEY_ID", "k1"),
		JWTKeyIDPrevious:  getEnvOrDefault("APP_JWT_KEY_ID_PREVIOUS", "prev"),

		// IP allowlisting
		AdminIPAllowlist: parseStringSlice("ADMIN_IP_ALLOWLIST"),

		// CORS
		CORSAllowedOrigins: parseStringSlice("CORS_ALLOWED_ORIGINS"),

		// Market data: providers + transport
		// MARKET_DATA_PROVIDER is honoured as an alias for
		// MARKET_DATA_PRIMARY_PROVIDER so the env example aligns with the
		// requested variable name. Explicit primary wins when both are set.
		MarketDataPrimaryProvider:  strings.ToLower(getEnvOrDefault("MARKET_DATA_PRIMARY_PROVIDER", getEnvOrDefault("MARKET_DATA_PROVIDER", "alpha_vantage"))),
		MarketDataFallbackProvider: strings.ToLower(getEnvOrDefault("MARKET_DATA_FALLBACK_PROVIDER", "yahoo")),
		// MARKET_DATA_TIMEOUT_MS is the new operator-facing knob; the
		// legacy MARKET_DATA_HTTP_TIMEOUT_SECONDS still wins when set so
		// existing deployments don't change behaviour.
		MarketDataHTTPTimeout: marketDataHTTPTimeout(),
		MarketDataCacheTTL:    time.Duration(parseInt("MARKET_DATA_CACHE_TTL_SECONDS", 900)) * time.Second,

		// Intraday valuation on the Holdings page.
		MarketDataStaleAfter:      time.Duration(parseInt("MARKET_DATA_STALE_AFTER_MINUTES", 15)) * time.Minute,
		MarketDataRefreshInterval: time.Duration(parseInt("MARKET_DATA_REFRESH_INTERVAL_MINUTES", 15)) * time.Minute,
		TwelveDataAPIKey:          os.Getenv("TWELVE_DATA_API_KEY"),

		// Market data: Alpha Vantage
		AlphaVantageAPIKey:          os.Getenv("ALPHA_VANTAGE_API_KEY"),
		AlphaVantageRateLimitPerMin: parseInt("ALPHA_VANTAGE_RATE_LIMIT_PER_MIN", 5),
		AlphaVantageRateLimitPerDay: parseInt("ALPHA_VANTAGE_RATE_LIMIT_PER_DAY", 500),
		AlphaVantageBaseURL:         getEnvOrDefault("ALPHA_VANTAGE_BASE_URL", "https://www.alphavantage.co"),
		AlphaVantageTimeout:         parseDuration("ALPHA_VANTAGE_TIMEOUT", "30s"),

		// Market data: ingestion scheduler
		IngestionScheduleCron: getEnvOrDefault("INGESTION_SCHEDULE_CRON", "0 18 * * 1-5"),

		// Investment: valuation staleness thresholds
		ValuationStaleThresholdsByAssetClass: parseStaleThresholds(
			"VALUATION_STALE_THRESHOLDS",
			"EQUITY=24h,FIXED_INCOME=72h,FUND=24h,ETF=24h,CASH=720h,ALTERNATIVE=720h,DERIVATIVE=24h",
		),

		// Chat
		LLMProvider:          strings.ToLower(getEnvOrDefault("LLM_PROVIDER", "anthropic")),
		AnthropicAPIKey:      os.Getenv("ANTHROPIC_API_KEY"),
		AnthropicModel:       getEnvOrDefault("ANTHROPIC_MODEL", "claude-haiku-4-5-20251001"),
		ChatMaxTokensPerTurn: parseInt("CHAT_MAX_TOKENS_PER_TURN", 1024),
		ChatWriteEnabled:     parseBool("CHAT_WRITE_ENABLED", false),
		ChatMCPServersConfig: os.Getenv("CHAT_MCP_SERVERS_CONFIG"),
		ChatMCPIMSBin:        os.Getenv("CHAT_MCP_IMS_BIN"),
		IMSAPIBaseURL:        getEnvOrDefault("IMS_API_BASE_URL", "http://localhost:8080/api/v1"),
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

	if cfg.AlphaVantageAPIKey == "" && cfg.Env != "development" && cfg.Env != "test" {
		return nil, fmt.Errorf("ALPHA_VANTAGE_API_KEY is required in non-development environments")
	}

	return cfg, nil
}

// MarshalJSON / String redactions for config: AppConfig has no String method
// today; if one is added, ensure JWTSecret, JWTSecretPrevious, DBPassword,
// MFAEncryptionKey, RedisPassword, and AlphaVantageAPIKey are NEVER emitted.

// marketDataHTTPTimeout reads the per-call provider HTTP timeout, preferring
// MARKET_DATA_TIMEOUT_MS when set (the operator-facing knob from the spec)
// and falling back to the legacy MARKET_DATA_HTTP_TIMEOUT_SECONDS / default.
func marketDataHTTPTimeout() time.Duration {
	if ms := parseInt("MARKET_DATA_TIMEOUT_MS", 0); ms > 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return time.Duration(parseInt("MARKET_DATA_HTTP_TIMEOUT_SECONDS", 15)) * time.Second
}

// parseStaleThresholds parses a comma-separated list of CODE=DURATION pairs
// (e.g. "EQUITY=24h,FIXED_INCOME=72h") into a map. Unparseable entries fall
// back to the default-string values silently — operability metrics are
// best-effort.
func parseStaleThresholds(key, defaultVal string) map[string]time.Duration {
	val := getEnvOrDefault(key, defaultVal)
	out := map[string]time.Duration{}
	for _, part := range strings.Split(val, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		eq := strings.IndexByte(part, '=')
		if eq <= 0 || eq == len(part)-1 {
			continue
		}
		code := strings.ToUpper(strings.TrimSpace(part[:eq]))
		dur, err := time.ParseDuration(strings.TrimSpace(part[eq+1:]))
		if err != nil {
			continue
		}
		out[code] = dur
	}
	return out
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
