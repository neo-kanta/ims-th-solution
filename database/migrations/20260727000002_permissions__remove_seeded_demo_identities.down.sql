-- =============================================================================
-- Rollback: documented no-op.
-- =============================================================================
-- Re-activating the demo identities and re-granting the seed-created access
-- this migration removes is exactly the production hazard it exists to close
-- (see the up.sql header and docs/MANAGER/MEMORY.md, "Financial and Safety
-- Rules": production migration/bootstrap paths must never assign privileges
-- or data scope to named demo identities). Rolling back this migration must
-- not restore that state.
--
-- Development/test parity for ben, green, and neo is unaffected by this
-- migration or its rollback: it is provided by the demo-only seeds under
-- database/seeds/demo/, which recreate the identities (INSERT ... ON CONFLICT
-- DO UPDATE, which also flips is_active back to true) and their grants
-- whenever APP_ENV is development or test, independently of migration state.
-- =============================================================================

BEGIN;

-- Intentionally empty: re-granting removed demo-identity access and
-- reactivating deactivated demo accounts is not a safe or supported rollback
-- action.

COMMIT;
