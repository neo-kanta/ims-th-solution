-- =============================================================================
-- Investment module — cash-request ownership consistency + terminal-state guard
-- =============================================================================
-- IMS-MERGE-BLOCKERS G2 item 7 (FUND_OPTIONAL ownership + terminal-state
-- constraints).
--
-- 1. Ownership consistency: investment__portfolio_cash_requests currently has
--    independent single-column FKs on portfolio_id and fund_id, so nothing at
--    the database level stops a row from claiming a fund_id that does not
--    actually belong to its portfolio_id (docs/MANAGER/MEMORY.md: "composite
--    foreign keys must prevent a row from claiming a fund different from its
--    portfolio's fund"). This mirrors the exact composite FK already added to
--    investment__decisions/executions/trade_confirmations by
--    20260703000001 (FOREIGN KEY (portfolio_id, fund_id) REFERENCES
--    investment__portfolios (id, fund_id), backed by that migration's
--    uq_inv_portfolios_id_fund UNIQUE (id, fund_id)) — no new unique
--    constraint is needed here, it already exists. The application never
--    supplies a client-chosen fund_id (post_transaction_cash_approval.go sets
--    FundID from the resolved portfolio's own FundID), so this is defense in
--    depth, not a fix for an observed bad row. Fund-optional aware: Postgres'
--    default MATCH SIMPLE means the composite FK is silently satisfied
--    whenever fund_id IS NULL (a fund-less submission) — only a fund-BOUND
--    row's (portfolio_id, fund_id) pair is actually checked against
--    investment__portfolios, which is exactly the fund-optional semantics
--    already used by the sibling trading tables.
--
-- 2. Terminal-state guard: a CHECK constraint cannot compare OLD vs NEW row
--    values, so the table's existing chk_inv_cash_req_status CHECK (status IN
--    (...)) cannot by itself stop a row from transitioning OUT of a terminal
--    status (APPROVED/REJECTED/CANCELLED) — e.g. an approve-then-cancel race,
--    a duplicate approval callback re-deciding an already-decided request, or
--    a future bug re-opening a terminal row. The application already
--    row-locks (GetForUpdate) and re-checks status before every transition,
--    but this trigger is a DB-level backstop that holds even against a bug or
--    a direct SQL statement outside the application's own guard. It ONLY
--    blocks a CHANGE of status away from an already-terminal value — a benign
--    re-save that leaves status unchanged (e.g. correcting memo) is
--    unaffected, and the PENDING -> {APPROVED,REJECTED,CANCELLED} transition
--    itself is unaffected (OLD.status = 'PENDING' is not in the terminal set).
-- =============================================================================

-- 1. Ownership consistency (composite FK; NULL fund_id is fund-optional-safe).
ALTER TABLE investment__portfolio_cash_requests
    ADD CONSTRAINT fk_inv_cash_req_portfolio_fund
        FOREIGN KEY (portfolio_id, fund_id)
        REFERENCES investment__portfolios (id, fund_id);

-- 2. Terminal-state guard.
CREATE OR REPLACE FUNCTION inv_cash_req_prevent_terminal_status_change() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IN ('APPROVED', 'REJECTED', 'CANCELLED')
       AND NEW.status IS DISTINCT FROM OLD.status THEN
        RAISE EXCEPTION
            'cash request % is in terminal status % and cannot transition to %',
            OLD.id, OLD.status, NEW.status
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_inv_cash_req_terminal_status_immutable
    BEFORE UPDATE ON investment__portfolio_cash_requests
    FOR EACH ROW EXECUTE FUNCTION inv_cash_req_prevent_terminal_status_change();

COMMENT ON FUNCTION inv_cash_req_prevent_terminal_status_change() IS
    'DB-level backstop: once a cash request reaches APPROVED/REJECTED/CANCELLED, its status can never change again, even outside the application''s own row-lock/re-check guard.';
