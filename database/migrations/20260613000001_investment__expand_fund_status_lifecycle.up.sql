-- =============================================================================
-- Investment Module — Fund status lifecycle expansion  (Phase 1 of 2)
-- =============================================================================
-- Adds DRAFT, PENDING_APPROVAL, and REJECTED to the fund status CHECK constraint
-- so funds can follow the full onboarding lifecycle:
--   DRAFT → PENDING_APPROVAL → ACTIVE → SUSPENDED / CLOSED / REJECTED
--
-- Two-phase strategy:
--   Phase 1 (this migration): expand CHECK to allow both old and new values.
--   Phase 2 (cleanup migration, separate PR): remove any values that have been
--   fully migrated away from — run only after code + frontend are updated.
--
-- The column DEFAULT remains 'ACTIVE' for backward compat with any tooling that
-- inserts a fund row without an explicit status. The new lifecycle-aware command
-- handler (CreateFundWithLifecycle) will always pass 'DRAFT' explicitly.
-- =============================================================================

ALTER TABLE investment__funds
    DROP CONSTRAINT chk_inv_funds_status;

ALTER TABLE investment__funds
    ADD CONSTRAINT chk_inv_funds_status
        CHECK (status IN (
            'DRAFT',
            'PENDING_APPROVAL',
            'ACTIVE',
            'SUSPENDED',
            'CLOSED',
            'REJECTED'
        ));
