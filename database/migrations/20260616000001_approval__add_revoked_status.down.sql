-- =============================================================================
-- Approval Module — remove REVOKED request status (rollback)
-- =============================================================================
-- WARNING: This will fail if any rows currently have status = 'REVOKED'.
-- Ensure all revoked requests are updated before running down.
-- =============================================================================

ALTER TABLE approval__requests
    DROP CONSTRAINT chk_approval_request_status,
    ADD CONSTRAINT chk_approval_request_status CHECK (status IN (
        'DRAFT',
        'SUBMITTED',
        'PENDING_APPROVAL',
        'APPROVED',
        'REJECTED',
        'CANCELLED',
        'WITHDRAWN'
    ));
