package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultAdminPasswordHash = "$2a$12$RlJ58G8tTbtR8.xKogKewOgRjs0RcGzW6S0JTmZbiYUJplGPnZO1C"

// EnsureSecureBootstrap blocks startup in non-development environments if an insecure default admin remains.
func EnsureSecureBootstrap(ctx context.Context, pool *pgxpool.Pool, env string) error {
	if strings.EqualFold(env, "development") || strings.EqualFold(env, "test") {
		return nil
	}

	var hasDefaultAdmin bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM iam_users
			WHERE username = 'admin'
			  AND password_hash = $1
			  AND deleted_at IS NULL
		)
	`, defaultAdminPasswordHash).Scan(&hasDefaultAdmin)
	if err != nil {
		return fmt.Errorf("checking secure bootstrap state: %w", err)
	}
	if hasDefaultAdmin {
		return fmt.Errorf("insecure bootstrap blocked: default admin credentials are still present")
	}
	return nil
}
