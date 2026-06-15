-- =============================================================================
-- Workflow Module — Expand workflow state values  (Phase 1 of 2)
-- =============================================================================
-- Adds INVESTMENT_DAY_STARTED and MANAGER_APPROVED_END_OF_DAY to the workflow
-- state CHECK constraints. The existing DAY_OPEN and MANAGER_APPROVED values
-- are RETAINED in Phase 1 to allow a safe two-phase migration:
--
--   Phase 1 (this migration):
--     Expand CHECK to allow both old and new state names.
--     Deploy new Go code that writes INVESTMENT_DAY_STARTED / MANAGER_APPROVED_END_OF_DAY.
--     Regenerate OpenAPI types.
--
--   Phase 2 (cleanup migration, separate PR):
--     1. UPDATE workflow__day_states
--            SET current_state = 'INVESTMENT_DAY_STARTED'
--          WHERE current_state = 'DAY_OPEN';
--        UPDATE workflow__day_states
--            SET current_state = 'MANAGER_APPROVED_END_OF_DAY'
--          WHERE current_state = 'MANAGER_APPROVED';
--        (similar UPDATE for workflow__transition_log from_state / to_state)
--     2. Drop DAY_OPEN and MANAGER_APPROVED from CHECK constraints.
--
-- Note: TRADING_OPEN is intentionally excluded (per architecture decision:
-- INVESTMENT_DAY_STARTED covers the full open-for-trading window; TRADING_OPEN
-- may be added later as a fund-level optional intermediate step).
-- =============================================================================

-- ── workflow__day_states ────────────────────────────────────────────────────

ALTER TABLE workflow__day_states
    DROP CONSTRAINT chk_wf_day_state;

ALTER TABLE workflow__day_states
    ADD CONSTRAINT chk_wf_day_state CHECK (current_state IN (
        'NOT_STARTED',
        'DAY_OPEN',                       -- Phase 1: kept for backward compat
        'INVESTMENT_DAY_STARTED',         -- canonical new name
        'MANAGER_APPROVED',               -- Phase 1: kept for backward compat
        'MANAGER_APPROVED_END_OF_DAY',    -- canonical new name
        'TRANSACTION_CLOSED',
        'ACCOUNTING_CLOSED'
    ));

-- ── workflow__transition_log ────────────────────────────────────────────────

ALTER TABLE workflow__transition_log
    DROP CONSTRAINT chk_wf_transition_from_state;

ALTER TABLE workflow__transition_log
    ADD CONSTRAINT chk_wf_transition_from_state CHECK (from_state IN (
        'NOT_STARTED',
        'DAY_OPEN',
        'INVESTMENT_DAY_STARTED',
        'MANAGER_APPROVED',
        'MANAGER_APPROVED_END_OF_DAY',
        'TRANSACTION_CLOSED',
        'ACCOUNTING_CLOSED'
    ));

ALTER TABLE workflow__transition_log
    DROP CONSTRAINT chk_wf_transition_to_state;

ALTER TABLE workflow__transition_log
    ADD CONSTRAINT chk_wf_transition_to_state CHECK (to_state IN (
        'NOT_STARTED',
        'DAY_OPEN',
        'INVESTMENT_DAY_STARTED',
        'MANAGER_APPROVED',
        'MANAGER_APPROVED_END_OF_DAY',
        'TRANSACTION_CLOSED',
        'ACCOUNTING_CLOSED'
    ));
