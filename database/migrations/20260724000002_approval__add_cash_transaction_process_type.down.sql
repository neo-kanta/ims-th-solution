-- =============================================================================
-- Rollback: Approval Module — Remove PORTFOLIO_CASH_TRANSACTION process type
-- =============================================================================
-- WARNING: This rollback will fail if any approval__process_configs row has
-- process_type = 'PORTFOLIO_CASH_TRANSACTION'. Delete those rows first.
-- =============================================================================

ALTER TABLE approval__process_configs
    DROP CONSTRAINT chk_approval_process_type;

ALTER TABLE approval__process_configs
    ADD CONSTRAINT chk_approval_process_type CHECK (process_type IN (
        'INVESTMENT_ANALYSIS_REPORT',
        'INVESTMENT_DECISION',
        'INVESTMENT_CANCELLATION',
        'WORKFLOW_OPERATION',
        'LEAVE_REQUEST',
        'LEAVE_CANCELLATION',
        'DELEGATION_REQUEST',
        'PORTFOLIO_ONBOARDING',
        'COMPLIANCE_RELEASE'
    ));
