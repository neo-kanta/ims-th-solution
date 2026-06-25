-- =============================================================================
-- Investment Module — Add PENDING_COMPLIANCE_RELEASE decision status
-- =============================================================================
-- Extends the investment__decisions.status CHECK constraint to include
-- 'PENDING_COMPLIANCE_RELEASE'. A decision enters this state when the IRG
-- pre-trade check returns BLOCK with all breaches marked overridable — instead
-- of an immediate rejection, a COMPLIANCE_RELEASE approval is created and the
-- decision waits for a compliance officer to release it. Once released, the
-- decision is automatically re-submitted to the INVESTMENT_DECISION approval
-- engine.
-- =============================================================================

ALTER TABLE investment__decisions
    DROP CONSTRAINT chk_inv_decision_status;

ALTER TABLE investment__decisions
    ADD CONSTRAINT chk_inv_decision_status CHECK (status IN (
        'DRAFT',
        'PENDING_APPROVAL',
        'APPROVED',
        'REJECTED',
        'CANCELLED',
        'READY_FOR_EXECUTION',
        'EXECUTED',
        'PENDING_COMPLIANCE_RELEASE'
    ));
