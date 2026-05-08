package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// catalogExecutor is the minimal subset of pgxpool.Pool required by the
// permission-catalog upsert. Defining it here keeps the seeder unit-testable
// with a recording fake.
type catalogExecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// catalogEntry is the result of aggregating one PermissionDefinition with
// the module name returned by its owning catalog.
type catalogEntry struct {
	Code        string
	Module      string
	Name        string
	Description string
}

// upsertPermissionCatalogSQL is the canonical UPSERT used by the seeder. It
// overwrites name / module / description on every run so a refactor in any
// module's permission package propagates to the DB on the next seed pass.
const upsertPermissionCatalogSQL = `
INSERT INTO permissions_function_definitions (code, module, name, description)
VALUES ($1, $2, $3, $4)
ON CONFLICT (code) DO UPDATE SET
    module = EXCLUDED.module,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW()`

// aggregateCatalogs flattens a slice of catalogs into a deterministic list of
// entries, returning a typed error when two modules declare the same code.
//
// Stable order: by code ascending. The seeder relies on this for predictable
// SQL execution ordering during tests.
func aggregateCatalogs(catalogs []contract.PermissionCatalog) ([]catalogEntry, error) {
	byCode := make(map[string]catalogEntry, 64)

	for _, cat := range catalogs {
		if cat == nil {
			continue
		}
		module := cat.Module()
		if module == "" {
			return nil, fmt.Errorf("permission catalog has empty module name")
		}
		for _, def := range cat.Permissions() {
			if def.Code == "" {
				return nil, fmt.Errorf("permission catalog %q declared a definition with empty code", module)
			}
			if def.Name == "" {
				return nil, fmt.Errorf("permission catalog %q declared code %q with empty name", module, def.Code)
			}
			if existing, dup := byCode[def.Code]; dup && existing.Module != module {
				return nil, fmt.Errorf(
					"permission code %q declared by both module %q and module %q",
					def.Code, existing.Module, module,
				)
			}
			byCode[def.Code] = catalogEntry{
				Code:        def.Code,
				Module:      module,
				Name:        def.Name,
				Description: def.Description,
			}
		}
	}

	entries := make([]catalogEntry, 0, len(byCode))
	for _, entry := range byCode {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Code < entries[j].Code })
	return entries, nil
}

// upsertPermissionCatalog walks the aggregated entries and runs the
// canonical upsert against the executor. Each row is upserted in its own
// statement so a single bad row reports a clear error instead of failing
// the whole batch silently.
func upsertPermissionCatalog(
	ctx context.Context,
	exec catalogExecutor,
	catalogs []contract.PermissionCatalog,
) error {
	entries, err := aggregateCatalogs(catalogs)
	if err != nil {
		return fmt.Errorf("aggregating permission catalogs: %w", err)
	}

	for _, entry := range entries {
		if _, err := exec.Exec(
			ctx,
			upsertPermissionCatalogSQL,
			entry.Code, entry.Module, entry.Name, entry.Description,
		); err != nil {
			return fmt.Errorf("upserting permission code %q: %w", entry.Code, err)
		}
	}
	return nil
}
