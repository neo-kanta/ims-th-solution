-- =============================================================================
-- Approval Module — Add PORTFOLIO_ONBOARDING process type
-- =============================================================================
-- Extends the approval__process_configs.process_type CHECK constraint to
-- include 'PORTFOLIO_ONBOARDING', enabling the approval engine to govern
-- the fund/portfolio onboarding lifecycle (DRAFT → PENDING_APPROVAL → ACTIVE).
--
-- The approval__requests table has no subject_type CHECK — the domain layer
-- validates subject types, so no schema change is required there.
--
-- After this migration, an approval process config of type PORTFOLIO_ONBOARDING
-- must be seeded (see database/seeds/) before any fund or portfolio can be
-- submitted for approval.
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
        'PORTFOLIO_ONBOARDING'
    ));
