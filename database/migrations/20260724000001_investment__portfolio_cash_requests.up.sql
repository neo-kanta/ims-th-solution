-- =============================================================================
-- Investment module — LIVE cash-transaction approval requests
-- =============================================================================
-- IMS-PORTFOLIO-FUND-OPTIONAL Stage 2.
--
-- investment__portfolio_transactions is APPEND-ONLY (RULEs
-- no_update_/no_delete_inv_portfolio_transactions rewrite UPDATE/DELETE to
-- NOTHING). A LIVE cash movement (CASH_IN/CASH_OUT/FEE/DIVIDEND) must therefore
-- NOT be written as a real ledger row while it is pending approval — it needs
-- its own MUTABLE staging entity that supports status transitions. The real
-- ledger row is materialized only after approval, inside the approval callback.
--
-- This table is intentionally MUTABLE — do NOT add append-only RULEs/triggers.
--
-- Fund-optional aware: fund_id is NULLABLE (mirrors the trading tables made
-- fund-less in 20260723000001 / 20260723000003). The data-scope key falls back
-- to portfolio_id when there is no fund.
-- =============================================================================

CREATE TABLE investment__portfolio_cash_requests (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id          UUID          NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    -- NULL for a fund-less portfolio; the scope fallback is portfolio_id.
    fund_id               UUID          REFERENCES investment__funds(id) ON DELETE RESTRICT,

    transaction_type      VARCHAR(20)   NOT NULL,
    -- amount is always positive; the sign of the cash impact is conveyed by
    -- transaction_type at materialization time (CASH_IN/DIVIDEND add,
    -- CASH_OUT/FEE subtract), mirroring computeNetAmount in the post pipeline.
    amount                DECIMAL(28,8)  NOT NULL,
    currency              CHAR(3)       NOT NULL,
    fees                  DECIMAL(28,8)  NOT NULL DEFAULT 0,
    value_date            DATE          NOT NULL,
    memo                  TEXT,

    status                VARCHAR(20)   NOT NULL DEFAULT 'PENDING',

    -- approval__requests.id returned by SubmitForApproval.
    approval_request_id   UUID,
    -- investment__portfolio_transactions.id set once materialized; also the
    -- idempotency guard for duplicate approval callbacks.
    resulting_txn_id      UUID          REFERENCES investment__portfolio_transactions(id) ON DELETE RESTRICT,

    submitted_by          UUID          NOT NULL REFERENCES iam_users(id),
    submitted_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    decided_by            UUID          REFERENCES iam_users(id),
    decided_at            TIMESTAMPTZ,

    version               INTEGER       NOT NULL DEFAULT 1,

    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    created_by            UUID          NOT NULL REFERENCES iam_users(id),
    updated_by            UUID          NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_cash_req_type
        CHECK (transaction_type IN ('CASH_IN','CASH_OUT','FEE','DIVIDEND')),

    CONSTRAINT chk_inv_cash_req_status
        CHECK (status IN ('PENDING','APPROVED','REJECTED','CANCELLED')),

    CONSTRAINT chk_inv_cash_req_amount_positive
        CHECK (amount > 0),

    CONSTRAINT chk_inv_cash_req_fees_non_negative
        CHECK (fees >= 0),

    CONSTRAINT chk_inv_cash_req_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_cash_req_version_positive
        CHECK (version >= 1)
);

-- Submitter read path (GET /cash-requests?status=) and portfolio-scoped listing.
CREATE INDEX idx_inv_cash_req_portfolio_status
    ON investment__portfolio_cash_requests (portfolio_id, status);

CREATE INDEX idx_inv_cash_req_submitted_by
    ON investment__portfolio_cash_requests (submitted_by);

-- One materialized ledger row per cash request.
CREATE UNIQUE INDEX uq_inv_cash_req_resulting_txn
    ON investment__portfolio_cash_requests (resulting_txn_id)
    WHERE resulting_txn_id IS NOT NULL;

CREATE TRIGGER trg_inv_cash_req_updated_at
    BEFORE UPDATE ON investment__portfolio_cash_requests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__portfolio_cash_requests IS
    'Mutable staging entity for LIVE-portfolio cash movements awaiting approval. The real append-only ledger row is materialized only after approval (see cash_request_approval_adapter.go). NOT append-only.';
COMMENT ON COLUMN investment__portfolio_cash_requests.amount IS
    'Positive magnitude; cash-impact sign is derived from transaction_type at materialization.';
COMMENT ON COLUMN investment__portfolio_cash_requests.resulting_txn_id IS
    'Set once the approved request is materialized into investment__portfolio_transactions. Duplicate-callback idempotency guard.';
