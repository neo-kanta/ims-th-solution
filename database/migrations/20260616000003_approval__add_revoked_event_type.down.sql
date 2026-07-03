-- =============================================================================
-- Approval Module — remove REVOKED from event type constraint (rollback)
-- =============================================================================
-- WARNING: This will fail if any rows in approval__events currently have
-- event_type = 'REVOKED'. Ensure all such rows are removed before rolling back.
-- =============================================================================

ALTER TABLE approval__events
    DROP CONSTRAINT IF EXISTS chk_approval_event_type,
    ADD CONSTRAINT chk_approval_event_type CHECK (event_type IN (
        'SUBMITTED','TASK_CREATED','APPROVED','REJECTED','CANCELLED','WITHDRAWN',
        'DELEGATED','STAGE_COMPLETED','REQUEST_COMPLETED'
    ));
