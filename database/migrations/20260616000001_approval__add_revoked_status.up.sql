-- =============================================================================
-- Approval Module — add REVOKED request status
-- =============================================================================
-- Extends the approval__requests status CHECK constraint to include REVOKED,
-- enabling a privileged user to undo a final approval and reopen the subject
-- for correction while preserving the full audit trail.
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
        'WITHDRAWN',
        'REVOKED'
    ));
