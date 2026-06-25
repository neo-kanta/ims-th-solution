-- =============================================================================
-- Investment Module — Add compliance_release_approval_request_id column
-- =============================================================================
-- Stores the COMPLIANCE_RELEASE approval request ID separately from the
-- INVESTMENT_DECISION approval request ID so both stages remain traceable via
-- the API without overwriting each other (see HIGH-1 fix).
-- =============================================================================

ALTER TABLE investment__decisions
    ADD COLUMN IF NOT EXISTS compliance_release_approval_request_id UUID NULL;
