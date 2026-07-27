package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

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

	// Demo seeds create or grant named demo identities (ben/green/neo) and demo
	// business data. They must only ever run on an explicit, verified opt-in:
	// setting INCLUDE_DEMO_SEEDS=true is honored ONLY when APP_ENV is explicitly
	// "development" or "test" (read directly via os.Getenv, not cfg.Env, which
	// already substitutes "development" for an unset APP_ENV — using cfg.Env here
	// would silently treat a missing APP_ENV as development and defeat the
	// fail-closed contract). Any other combination — including
	// INCLUDE_DEMO_SEEDS=true with APP_ENV unset, empty, or anything other than
	// development/test — is a hard configuration error, checked and rejected here
	// before any database connection or write.
	//
	// Without the opt-in (demoAllowed=false), seeding is reference-only: only the
	// seeds listed in database/seeds/reference_manifest.txt run. An unlisted file
	// under a demo/ or zz_demo/ directory is skipped quietly (the normal
	// production outcome); an unlisted file NOT in a demo location is a hard error
	// (the manifest is out of sync with the tree). A missing/misconfigured
	// manifest, or SEEDS_PATH pointed at a subdirectory that has no manifest, is
	// likewise a hard, non-zero failure rather than a silent seed-nothing (see
	// runSQLSeeds).
	rawAppEnv := os.Getenv("APP_ENV")
	includeDemoSeeds := strings.EqualFold(os.Getenv("INCLUDE_DEMO_SEEDS"), "true")
	demoAllowed, err := resolveDemoSeeding(rawAppEnv, includeDemoSeeds)
	if err != nil {
		slog.Error("Invalid demo seed configuration", "error", err)
		os.Exit(1)
	}

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
	slog.Info("Running SQL seeds", "dir", seedsDir, "appEnv", rawAppEnv, "demoSeedsAllowed", demoAllowed)
	if err := runSQLSeeds(ctx, pool, seedsDir, demoAllowed); err != nil {
		slog.Error("Failed to run SQL seeds", "error", err)
		os.Exit(1)
	}

	slog.Info("Database seeded successfully")
}
