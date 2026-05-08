package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
	"github.com/neo-kanta/ims-th-solution/backend/platform/logging"
)

const defaultSeedsDir = "../database/seeds"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger := logging.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	catalogs := defaultPermissionCatalogs()
	if err := upsertPermissionCatalog(ctx, pool, catalogs); err != nil {
		slog.Error("Failed to upsert permission catalog", "error", err)
		os.Exit(1)
	}
	slog.Info("Permission catalog upserted", "modules", len(catalogs))

	seedsDir := os.Getenv("SEEDS_PATH")
	if seedsDir == "" {
		seedsDir = defaultSeedsDir
	}
	if err := runSQLSeeds(ctx, pool, seedsDir); err != nil {
		slog.Error("Failed to run SQL seeds", "error", err)
		os.Exit(1)
	}

	slog.Info("Database seeded successfully")
}
