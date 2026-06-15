-- =============================================================================
-- Workflow — accounting / NAV date separated from business date
-- =============================================================================
-- The investment day's business_date is the trading calendar date for the
-- contract. The accounting_date (a.k.a. NAV date) is the date the close was
-- posted to GL / NAV — typically the same business day in steady-state but
-- distinguishable when an operator backdates accounting close, when a
-- rollback re-posts to a different NAV cycle, or when accounting runs on a
-- delayed schedule.
--
-- accounting_date is set by close_accounting; prev_accounting_date stashes
-- the previous accounting_date during a rollback so audit can reconstruct
-- the accounting cycle that was reversed.
-- =============================================================================

ALTER TABLE workflow__day_states
    ADD COLUMN IF NOT EXISTS accounting_date      DATE,
    ADD COLUMN IF NOT EXISTS prev_accounting_date DATE;

-- Accounting date may not be earlier than the business date — accounting
-- cannot be posted before the day it covers.
ALTER TABLE workflow__day_states
    ADD CONSTRAINT chk_wf_accounting_date_after_business_date
        CHECK (accounting_date IS NULL OR accounting_date >= business_date);

COMMENT ON COLUMN workflow__day_states.accounting_date IS
    'NAV / accounting posting date for the closed day. Set at CLOSE_ACCOUNTING; cleared by rollback. May equal or be later than business_date.';
COMMENT ON COLUMN workflow__day_states.prev_accounting_date IS
    'Last accounting_date prior to the current rollback cycle. Auditors use it to reconstruct the original NAV cycle.';
