-- =============================================================================
-- Rollback: Investment Module — Fund status lifecycle expansion (Phase 1)
-- =============================================================================
-- WARNING: This rollback will fail if any fund row has status
-- 'DRAFT', 'PENDING_APPROVAL', or 'REJECTED'. Purge or migrate those rows
-- before running down.
-- =============================================================================

ALTER TABLE investment__funds
    DROP CONSTRAINT chk_inv_funds_status;

ALTER TABLE investment__funds
    ADD CONSTRAINT chk_inv_funds_status
        CHECK (status IN ('ACTIVE', 'SUSPENDED', 'CLOSED'));
