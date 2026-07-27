# Database Seeds

Seeds populate reference/catalog data in every environment and, only on an
explicit opt-in, demo data after migrations are applied.

## Layout

```text
database/seeds/
├── 001_initial_seed.sql                      # placeholder / common
├── 002_workflow_seed.sql                     # workflow groups + grants
├── 003_workflow_settings_seed.sql            # workflow scheduler settings
├── 004_investment_process_assignment_seed.sql  # process step catalog (reference)
├── 012_approval_process_seed.sql             # approval groups/process/stages (reference)
├── 018_investment_decision_execution_permission_seed.sql # decision/execution role catalog (reference)
├── investment/                               # Phase 0 investment reference data
│   ├── 01_asset_classes.sql
│   ├── ...
│   └── 07_investment_styles.sql
├── demo/                                     # development/test ONLY — see below
│   ├── 001_investment_process_assignment_demo_seed.sql
│   ├── 002_approval_demo_seed.sql
│   ├── 003_approval_portfolio_onboarding_demo_seed.sql
│   ├── 004_ben_workflow_approver_seed.sql
│   ├── 005_investment_decision_operator_ben_membership_seed.sql
│   ├── 006_ben_investment_decision_approver_seed.sql
│   ├── 007_ben_operation_page_access_seed.sql
│   ├── 008_green_ben_workflow_permission_parity_seed.sql
│   └── 009_workflow_scheduler_contracts_demo_seed.sql
└── zz_demo/                                   # development/test ONLY — business-data demo (see zz_demo/README.md)
```

## Reference vs. demo, and the production gate

`backend/cmd/seed` (see `sql_seeds.go`) classifies every `*.sql` file it finds
under `database/seeds` against a checked-in **reference manifest**
(`database/seeds/reference_manifest.txt`) — an explicit allow-list, NOT
directory-name text matching:

- **Reference** — a file whose path (relative to the resolved canonical seed
  root) is listed in `reference_manifest.txt`. Today that is the 15 top-level
  numbered files plus `investment/01..07`. This is catalog/config data
  (permission definitions, role shells, compliance rules, market-data
  securities, process/approval wiring, cash-transaction process config) that
  every environment, including production, needs to boot. It never creates or
  grants a named identity — only catalog rows and, where a working approver is
  required in production, the bootstrap `admin` account.
- **Demo** — any file NOT listed in the manifest. The demo seeds under
  `database/seeds/demo/**` and `database/seeds/zz_demo/**` create the named demo
  users (`ben`/`green`/`neo`), assign them to role catalogs, and seed demo
  business data (funds, positions, prices, workflow history, compliance
  breaches, and the synthetic `IMS-DEMO-*` scheduler contracts). None of these
  are in the manifest, so none run in production.

Classification is by the path **relative to the resolved canonical seed root**
(the root is symlink-resolved via `EvalSymlinks`), so it does not matter where on
disk the repository lives — a repo checked out under a parent directory literally
named `demo`, a renamed demo directory, or a symlink/junction cannot change what
runs. Fail-closed behavior when demo seeding is NOT opted in (i.e. production):

- Only manifest-listed files run. An unlisted file under `demo/`/`zz_demo/` is
  skipped quietly; an unlisted file **not** in a demo location is a HARD ERROR
  (the manifest is out of sync with the tree).
- A **missing or empty** manifest, **zero** `*.sql` files, or any manifest-listed
  reference file that is **absent, not a regular file, or resolves (via a
  symlink/junction) outside the canonical seed root** is a HARD, non-zero
  failure — never a silent success. Pointing `SEEDS_PATH` at a subdirectory that
  has no manifest therefore fails loudly rather than running that subdirectory's
  files as if they were reference data.

### Enabling demo seeds requires an explicit opt-in

Demo seeding is OFF by default, in every `APP_ENV`, including `development`
and `test`. To run demo seeds, both of the following must hold:

1. `INCLUDE_DEMO_SEEDS=true` is set.
2. `APP_ENV` is **explicitly** `development` or `test`.

If `INCLUDE_DEMO_SEEDS=true` is set but `APP_ENV` is anything else —
including unset/empty — `cmd/seed` exits non-zero **before connecting to the
database or seeding anything**. An unset `APP_ENV` is never treated as an
implicit `development` for this decision, even though `platform/config`
defaults `cfg.Env` to `"development"` for unrelated purposes; the demo-seed
gate reads the raw `APP_ENV` environment variable directly so a missing
`APP_ENV` cannot silently unlock demo seeding.

Without the opt-in (the default), seeding is reference-only and **always
succeeds**, in every environment, including production — demo directories
existing on disk are silently skipped and never cause a non-zero exit. With
the opt-in verified for `development`/`test`, every file — reference and demo
— executes in the same lexicographic order as before (unchanged local/CI
seeding behavior once the opt-in is set).

Some reference files previously also contained named-demo-user content (e.g.
the former `004_investment_process_assignment_seed.sql` created ben/green/neo,
and `012_approval_demo_seed.sql` assigned ben/green as approval-group
members). That content has been extracted into `demo/*`; the original
filenames now contain only the reference/catalog portion (`012` was renamed
to `012_approval_process_seed.sql` to reflect this).

### Production has no seeded scheduler contracts by design

`workflow__scheduler_contracts` is an explicitly temporary bridge table (see
`backend/internal/workflow/infrastructure/persistence/scheduler_repository.go`)
that the workflow scheduler reads to decide which contracts to open/close
each business day, until a real contract/fund master module owns active
contracts. The only seed that ever populated it
(`demo/009_workflow_scheduler_contracts_demo_seed.sql`, formerly the
always-run `005_workflow_scheduler_contracts_seed.sql`) inserts six synthetic
`IMS-DEMO-*` contracts. Letting synthetic contracts reach the production
scheduler was a defect, not a feature, so this file is now demo-only: a
production (or any non-opted-in) deployment seeds **zero** rows into
`workflow__scheduler_contracts`, and the scheduler correctly finds no
contracts to act on until real contract/fund data is wired in. No other
reference seed depends on these rows — `workflow__day_states.contract_id` (in
`002_workflow_seed.sql`) is a plain UUID column carrying a fixed legacy
placeholder, not a foreign key to this table — so moving this file does not
break reference seeding.

## Conventions

- All seeds are **idempotent** — every `INSERT` carries `ON CONFLICT ... DO NOTHING` (or `DO UPDATE` for the permission catalog) so re-running `make seed` leaves operator-edited rows alone.
- `make db-reset` chains `migrate-up` and `seed`, so seeds run after every migration.
- Reference files run in lexicographic order; subdirectory contents follow top-level files (e.g. `004_*.sql` before `investment/01_*.sql`, before `demo/*.sql`, before `zz_demo/*.sql`) when demo seeding is enabled.
- Seeds reference table names defined by **already-applied** migrations; if you introduce a table in a new migration, ship the matching seed in the same change.
- A reference seed must never depend on a `demo/` or `zz_demo/` row to satisfy a foreign key, since reference seeds must succeed on their own in production.

## Permission catalog

`cmd/seed/main.go` runs Go-side permission catalog upserts before applying SQL files. Each module that owns permission codes implements `contract.PermissionCatalog`; the seeder iterates the registered providers and writes / refreshes rows in `permissions_function_definitions`. The SQL seeds may then grant those codes to groups via `permissions_function_rights`.
