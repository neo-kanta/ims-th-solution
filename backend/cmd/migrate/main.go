package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

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

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
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

	if command == "up" {
		slog.Info("Running migrations UP...")
		err = m.Up()
	} else if command == "down" {
		slog.Info("Running migrations DOWN (1 step)...")
		err = m.Steps(-1)
	} else if command == "force" {
		if len(os.Args) != 3 {
			slog.Error("Force requires exactly one migration version", "usage", "migrate force <version>")
			os.Exit(1)
		}
		version, parseErr := strconv.Atoi(os.Args[2])
		if parseErr != nil || version < 0 {
			slog.Error("Invalid migration version", "version", os.Args[2])
			os.Exit(1)
		}
		slog.Warn("Forcing migration version; verify the database schema before continuing", "version", version)
		err = m.Force(version)
	} else {
		slog.Error("Invalid migration command", "command", command, "usage", "migrate [up|down|force <version>]")
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
