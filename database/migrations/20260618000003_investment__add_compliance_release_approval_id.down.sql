-- =============================================================================
-- Rollback: Investment Module — Remove compliance_release_approval_request_id
-- =============================================================================

ALTER TABLE investment__decisions
    DROP COLUMN IF EXISTS compliance_release_approval_request_id;
