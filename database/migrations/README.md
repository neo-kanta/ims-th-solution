# Database Migrations

This directory contains the SQL migrations for the IMS Thailand PostgreSQL schema.

The project uses the Go migration entrypoint in `backend/cmd/migrate/main.go`, and the common workflow is exposed through the repository `Makefile`.

## Naming Convention

Migration files follow this pattern:

```text
<timestamp>_<module>__<action>.up.sql
<timestamp>_<module>__<action>.down.sql
```

Examples:

```text
20260408102212_iam__add_security_features.up.sql
20260408102212_iam__add_security_features.down.sql
```

## Common Commands

Run these from the repository root.

### Apply all pending migrations

```bash
make migrate-up
```

### Roll back the latest migration

```bash
make migrate-down
```

### Create a new migration pair

```bash
make migrate-new module=iam name=add_mfa_enrollment
```

### Reset the local database

```bash
make db-reset
```

This will recreate the local PostgreSQL container, run all migrations, and then execute the seed step.

## Writing Good Migrations

- Always create both `up` and `down` files
- Keep each migration focused on one logical schema change
- Make `down` scripts fully reverse the `up` change where practical
- Prefer additive, forward-safe changes for production-bound schema evolution
- Test the migration on a local database before opening a PR

## Seeds

Development seed SQL is loaded separately from `database/seeds/`.

Notes:

- `make seed` runs the seed loader from `backend/cmd/seed/main.go`
- the current seed file is a placeholder template, not a guaranteed set of demo users or contracts

## Practical Workflow

1. Create the migration with `make migrate-new`
2. Fill in the `up` and `down` files
3. Run `make migrate-up`
4. If needed, test rollback with `make migrate-down`
5. Re-apply with `make migrate-up`

## Related Files

- `backend/cmd/migrate/main.go`
- `backend/platform/database/`
- `database/seeds/`
