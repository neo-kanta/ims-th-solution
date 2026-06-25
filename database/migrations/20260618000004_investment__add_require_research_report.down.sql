-- =============================================================================
-- Rollback: Investment Module — Remove require_research_report_for_decision
-- =============================================================================

ALTER TABLE investment__funds
    DROP COLUMN IF EXISTS require_research_report_for_decision;
