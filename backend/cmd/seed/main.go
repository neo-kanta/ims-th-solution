package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
	"github.com/neo-kanta/ims-th-solution/backend/platform/logging"
)

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

	seedsDir := os.Getenv("SEEDS_PATH")
	if seedsDir == "" {
		seedsDir = "../database/seeds"
	}
	files, err := filepath.Glob(filepath.Join(seedsDir, "*.sql"))
	if err != nil {
		slog.Error("Failed to list seed files", "error", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		slog.Info("No seed files found")
		return
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			slog.Error("Failed to read seed file", "file", file, "error", err)
			os.Exit(1)
		}

		slog.Info("Executing seed file", "file", file)
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			slog.Error("Failed to execute seed file", "file", file, "error", err)
			os.Exit(1)
		}
	}

	slog.Info("Database seeded successfully")
}
