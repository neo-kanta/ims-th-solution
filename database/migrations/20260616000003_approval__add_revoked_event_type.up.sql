-- =============================================================================
-- Approval Module — add REVOKED to event type constraint
-- =============================================================================
-- The chk_approval_event_type constraint on approval__events did not include
-- REVOKED, causing every RevokeRequest() call to fail at DB level after the
-- request status was already updated. This migration adds REVOKED to the
-- allowed event types so the full revoke flow (status update + event) succeeds
-- atomically.
-- =============================================================================

ALTER TABLE approval__events
    DROP CONSTRAINT IF EXISTS chk_approval_event_type,
    ADD CONSTRAINT chk_approval_event_type CHECK (event_type IN (
        'SUBMITTED','TASK_CREATED','APPROVED','REJECTED','CANCELLED','WITHDRAWN',
        'DELEGATED','STAGE_COMPLETED','REQUEST_COMPLETED','REVOKED'
    ));
