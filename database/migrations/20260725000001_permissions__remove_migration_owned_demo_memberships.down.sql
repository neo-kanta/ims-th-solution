-- =============================================================================
-- Forward cleanup: remove migration-owned demo-user group memberships
-- =============================================================================
-- Documented no-op. Re-creating the six named-demo-user (ben/green) group
-- memberships this migration removes is exactly the production/upgraded-
-- database hazard this migration exists to close (see the up.sql header and
-- docs/MANAGER/MEMORY.md, "Financial and Safety Rules": production migration/
-- bootstrap paths must never assign privileges to named demo identities).
-- Rolling back this migration must not restore that state.
--
-- Development/test parity for ben and green is unaffected by this migration
-- or its rollback: it is provided by the demo-only seeds under
-- database/seeds/demo/, which run independently of migration state whenever
-- APP_ENV is development or test.
-- =============================================================================

BEGIN;

-- Intentionally empty: re-granting removed demo-identity memberships is not a
-- safe or supported rollback action.

COMMIT;
