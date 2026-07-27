-- =============================================================================
-- Green/Ben permission-parity copy — retired (documented no-op)
-- =============================================================================
-- The up migration no longer assigns anything (see the up.sql header for the
-- full rationale: production/upgraded-database paths must never assign
-- privileges to named demo identities). There is nothing for this down
-- migration to reverse.
--
-- This file previously deleted the four Green membership rows the up
-- migration inserted (by fixed membership id). Those rows are no longer
-- inserted by the up migration, and any that already exist in an
-- already-upgraded database from a prior version of this migration are
-- removed by the separate forward-corrective migration
-- 20260725000001_permissions__remove_migration_owned_demo_memberships, not by
-- rollback of this migration.
-- =============================================================================

BEGIN;

-- Intentionally empty: no schema change to reverse.

COMMIT;
