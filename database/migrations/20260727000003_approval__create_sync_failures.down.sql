-- =============================================================================
-- Rollback: drop the sync-failure replay table.
-- =============================================================================

BEGIN;

DROP TABLE IF EXISTS approval__sync_failures;

COMMIT;
