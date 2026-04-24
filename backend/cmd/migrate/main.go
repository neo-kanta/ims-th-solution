package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "../database/migrations"
	}
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	dbURL := cfg.MigrationDSN()

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		slog.Error("Failed to create migrator", "error", err)
		os.Exit(1)
	}
	defer m.Close()

	if direction == "up" {
		slog.Info("Running migrations UP...")
		err = m.Up()
	} else if direction == "down" {
		slog.Info("Running migrations DOWN (1 step)...")
		err = m.Steps(-1)
	} else {
		slog.Error("Invalid direction", "direction", direction)
		os.Exit(1)
	}

	if err != nil {
		if err == migrate.ErrNoChange {
			slog.Info("No new migrations to apply")
		} else {
			slog.Error("Migration failed", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Info("Migrations successful")
	}
}
