package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
)

// NewRedisClient creates and verifies a Redis connection.
// Returns an error if the connection cannot be established (fail-fast).
func NewRedisClient(ctx context.Context, cfg *config.AppConfig) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
		PoolSize:     cfg.RedisPoolSize,
	}

	if cfg.RedisTLSEnabled {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("pinging redis at %s: %w", cfg.RedisAddr, err)
	}

	slog.Info("Redis connected", "addr", cfg.RedisAddr, "db", cfg.RedisDB)
	return client, nil
}

// RedisHealthCheck verifies the Redis connection is alive.
func RedisHealthCheck(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
