-- =============================================================================
-- Green/Ben permission-parity copy — retired (documented no-op)
-- =============================================================================
-- The original version of this migration copied the demo user "ben"'s
-- function-permission-group memberships (Fund Manager, Investment Decision
-- Operator, Investment Decision Approver, Investment Operation Page Access)
-- onto the demo user "green" by hardcoded user id, so both demo accounts
-- exercised the same workflows in development.
--
-- Production and other upgraded-database migration/bootstrap paths must never
-- assign privileges or data scope to named demo identities (see
-- docs/MANAGER/MEMORY.md, "Financial and Safety Rules"). Copying membership
-- for the specific users "ben" and "green" is exactly that, so the assignment
-- logic has been removed entirely rather than reworded.
--
-- The equivalent development/test-only membership copy now lives in
-- database/seeds/demo/008_green_ben_workflow_permission_parity_seed.sql,
-- which is gated by APP_ENV and only runs in development/test.
--
-- This migration is retained (rather than deleted) because it may already be
-- a recorded, applied version in some database's migration history; rewriting
-- migration history is unsafe. It is kept as an intentional no-op so
-- `migrate up` on a fresh database records this version without granting
-- anything. See 20260725000001_permissions__remove_migration_owned_demo_memberships
-- for the forward-corrective cleanup of any already-upgraded database that
-- applied the original version of this migration.
-- =============================================================================

BEGIN;

-- Intentionally empty: no named-identity assignment, no schema change.

COMMIT;
