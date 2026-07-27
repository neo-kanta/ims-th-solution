-- =============================================================================
-- Rollback: drop the terminal-state guard trigger/function and the ownership
-- composite FK.
-- =============================================================================

DROP TRIGGER IF EXISTS trg_inv_cash_req_terminal_status_immutable ON investment__portfolio_cash_requests;
DROP FUNCTION IF EXISTS inv_cash_req_prevent_terminal_status_change();

ALTER TABLE investment__portfolio_cash_requests
    DROP CONSTRAINT IF EXISTS fk_inv_cash_req_portfolio_fund;
