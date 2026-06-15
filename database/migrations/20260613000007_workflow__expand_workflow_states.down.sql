-- =============================================================================
-- Rollback: Workflow Module — Expand workflow state values (Phase 1)
-- =============================================================================
-- WARNING: This rollback will fail if any row has current_state,
-- from_state, or to_state equal to 'INVESTMENT_DAY_STARTED' or
-- 'MANAGER_APPROVED_END_OF_DAY'. Migrate those rows back to the old names
-- before running down.
-- =============================================================================

ALTER TABLE workflow__transition_log
    DROP CONSTRAINT chk_wf_transition_to_state;

ALTER TABLE workflow__transition_log
    ADD CONSTRAINT chk_wf_transition_to_state CHECK (to_state IN (
        'NOT_STARTED', 'DAY_OPEN', 'MANAGER_APPROVED',
        'TRANSACTION_CLOSED', 'ACCOUNTING_CLOSED'
    ));

ALTER TABLE workflow__transition_log
    DROP CONSTRAINT chk_wf_transition_from_state;

ALTER TABLE workflow__transition_log
    ADD CONSTRAINT chk_wf_transition_from_state CHECK (from_state IN (
        'NOT_STARTED', 'DAY_OPEN', 'MANAGER_APPROVED',
        'TRANSACTION_CLOSED', 'ACCOUNTING_CLOSED'
    ));

ALTER TABLE workflow__day_states
    DROP CONSTRAINT chk_wf_day_state;

ALTER TABLE workflow__day_states
    ADD CONSTRAINT chk_wf_day_state CHECK (current_state IN (
        'NOT_STARTED', 'DAY_OPEN', 'MANAGER_APPROVED',
        'TRANSACTION_CLOSED', 'ACCOUNTING_CLOSED'
    ));
