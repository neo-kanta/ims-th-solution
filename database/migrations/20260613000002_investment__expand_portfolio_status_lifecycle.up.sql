-- =============================================================================
-- Investment Module — Portfolio status lifecycle expansion  (Phase 1 of 2)
-- =============================================================================
-- Adds DRAFT, PENDING_APPROVAL, SUSPENDED, and REJECTED to the portfolio status
-- CHECK constraint. PAUSED is retained for backward compat during this phase.
-- The target lifecycle is:
--   DRAFT → PENDING_APPROVAL → ACTIVE → SUSPENDED / CLOSED / REJECTED
--
-- Two-phase strategy:
--   Phase 1 (this migration): expand CHECK; keep PAUSED alongside new SUSPENDED.
--   Phase 2 (cleanup migration, separate PR):
--     1. UPDATE investment__portfolios SET status = 'SUSPENDED' WHERE status = 'PAUSED';
--     2. Drop PAUSED from the CHECK constraint.
--   Run Phase 2 only after all application code and generated OpenAPI types no
--   longer reference 'PAUSED'.
-- =============================================================================

ALTER TABLE investment__portfolios
    DROP CONSTRAINT chk_inv_portfolios_status;

ALTER TABLE investment__portfolios
    ADD CONSTRAINT chk_inv_portfolios_status
        CHECK (status IN (
            'DRAFT',
            'PENDING_APPROVAL',
            'ACTIVE',
            'PAUSED',        -- kept for Phase 1 backward compat; removed in Phase 2
            'SUSPENDED',
            'CLOSED',
            'REJECTED'
        ));
