# Database Seeds

Seeds populate development and reference data after migrations are applied.

## Layout

```text
database/seeds/
├── 001_initial_seed.sql                      # placeholder / common
├── 002_workflow_seed.sql                     # workflow groups + grants
├── 003_workflow_settings_seed.sql            # workflow scheduler settings
├── 004_investment_process_assignment_seed.sql  # process group / assignment dev fixtures
├── 005_workflow_scheduler_contracts_seed.sql # scheduler dev contracts
└── investment/                               # Phase 0 investment reference data
    ├── 01_asset_classes.sql
    ├── 02_asset_subtypes.sql
    ├── 03_regions.sql
    ├── 04_countries.sql                      # ISO-3166-1 alpha-2 (inline VALUES)
    ├── 05_sectors.sql                        # GICS L1 + L2
    ├── 06_fund_categories.sql
    └── 07_investment_styles.sql
```

## Conventions

- All seeds are **idempotent** — every `INSERT` carries `ON CONFLICT ... DO NOTHING` (or `DO UPDATE` for the permission catalog) so re-running `make seed` leaves operator-edited rows alone.
- `make db-reset` chains `migrate-up` and `seed`, so seeds run after every migration.
- Files run in lexicographic order; subdirectory contents follow top-level files (e.g. `005_*.sql` before `investment/01_*.sql`).
- Seeds reference table names defined by **already-applied** migrations; if you introduce a table in a new migration, ship the matching seed in the same change.

## Permission catalog

`cmd/seed/main.go` runs Go-side permission catalog upserts before applying SQL files. Each module that owns permission codes implements `contract.PermissionCatalog`; the seeder iterates the registered providers and writes / refreshes rows in `permissions_function_definitions`. The SQL seeds may then grant those codes to groups via `permissions_function_rights`.
