-- =============================================================================
-- Rollback: Investment Module — Remove PENDING_COMPLIANCE_RELEASE status
-- =============================================================================
-- WARNING: This rollback will fail if any investment__decisions row has
-- status = 'PENDING_COMPLIANCE_RELEASE'. Update those rows first.
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
        'EXECUTED'
    ));
