-- =============================================================================
-- Approval Module — Add PORTFOLIO_CASH_TRANSACTION process type
-- =============================================================================
-- Extends the approval__process_configs.process_type CHECK constraint to
-- include 'PORTFOLIO_CASH_TRANSACTION', enabling the shared approval engine to
-- govern LIVE-portfolio cash movements (CASH_IN/CASH_OUT/FEE/DIVIDEND) before
-- they are materialized into the append-only ledger.
--
-- The approval__requests table has no subject_type CHECK — the domain layer
-- validates subject types (CASH_TRANSACTION), so no schema change is required
-- there.
--
-- After this migration, an approval process config of type
-- PORTFOLIO_CASH_TRANSACTION must be seeded (see
-- database/seeds/020_cash_transaction_process_seed.sql) before any cash request
-- can be submitted for approval.
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
        'COMPLIANCE_RELEASE',
        'PORTFOLIO_CASH_TRANSACTION'
    ));
