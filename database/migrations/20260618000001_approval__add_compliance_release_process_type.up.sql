-- =============================================================================
-- Approval Module — Add COMPLIANCE_RELEASE process and subject types
-- =============================================================================
-- Extends the approval__process_configs.process_type CHECK constraint to
-- include 'COMPLIANCE_RELEASE', enabling the shared approval engine to govern
-- the two-phase compliance release flow: when an investment decision is blocked
-- by an overridable IRG breach, a COMPLIANCE_RELEASE approval is created so a
-- compliance officer can release the breach before the decision continues to
-- the standard INVESTMENT_DECISION approval.
--
-- The approval__requests table has no subject_type CHECK — the domain layer
-- validates subject types, so no schema change is required there.
--
-- After this migration, an approval process config of type COMPLIANCE_RELEASE
-- must be seeded before any compliance release can be submitted for approval.
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
