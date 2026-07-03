-- =============================================================================
-- Rollback: Investment Module — Portfolio status lifecycle expansion (Phase 1)
-- =============================================================================
-- WARNING: This rollback will fail if any portfolio row has status
-- 'DRAFT', 'PENDING_APPROVAL', 'SUSPENDED', or 'REJECTED'. Purge or migrate
-- those rows before running down.
-- =============================================================================

ALTER TABLE investment__portfolios
    DROP CONSTRAINT chk_inv_portfolios_status;

ALTER TABLE investment__portfolios
    ADD CONSTRAINT chk_inv_portfolios_status
        CHECK (status IN ('ACTIVE', 'PAUSED', 'CLOSED'));
