// Package service holds the approval module's application services:
// ApprovalConfigService (configuration CRUD) and ApprovalRuntimeService
// (submit / approve / reject / withdraw / cancel + reads). Business rules such
// as maker-checker enforcement and stage transitions live here; transport
// stays thin and persistence stays SQL-only.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// withTransaction runs fn inside a pgx transaction, rolling back on error/panic.
func withTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("approval: beginning transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// nowUTC returns the current UTC time; injectable for tests via the service.
func nowUTC() time.Time { return time.Now().UTC() }
