-- Table: investment__portfolio_transactions
-- Source: 20260428000004_investment__create_ledger.up.sql,
--         updated by 20260429000100_investment__realised_pnl_and_valuation_line_delete_guard.up.sql,
--         updated by 20260430000001_investment__harden_immutability.up.sql,
--         updated by 20260703000001_investment__portfolio_v2_hardening.up.sql
-- Purpose: Append-only portfolio ledger. Portfolio is the operational/accounting source of truth.
-- Identity: Operational rows use portfolio_id; fund_id remains a legacy compatibility field
--           and must match the portfolio's fund_id while it exists.
CREATE TABLE investment__portfolio_transactions (
    -- Identity
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id             UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,

    -- Optional fund wrapper / legacy compatibility
    fund_id                  UUID         NOT NULL REFERENCES investment__funds(id)      ON DELETE RESTRICT,

    -- Instrument
    instrument_id            UUID         REFERENCES investment__instruments(id)         ON DELETE RESTRICT,

    -- Transaction economics
    transaction_type         VARCHAR(20)  NOT NULL,
    side                     VARCHAR(10),
    quantity                 DECIMAL(28,8),
    price                    DECIMAL(28,8),
    currency                 CHAR(3)      NOT NULL,
    gross_amount             DECIMAL(28,8) NOT NULL,
    fees                     DECIMAL(28,8) NOT NULL DEFAULT 0,
    net_amount               DECIMAL(28,8) NOT NULL,
    realised_pnl_base        DECIMAL(28,8) NOT NULL DEFAULT 0,
    fx_rate_to_base          DECIMAL(28,12),

    -- Dates
    business_date            DATE         NOT NULL,
    settlement_date          DATE,

    -- Source links
    source_decision_id       UUID,
    source_execution_id      UUID,
    reverses_transaction_id  UUID         REFERENCES investment__portfolio_transactions(id) ON DELETE RESTRICT,
    external_ref             VARCHAR(80),
    reason                   TEXT,

    -- Lifecycle / status
    status                   VARCHAR(20)  NOT NULL DEFAULT 'POSTED',

    -- Audit
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by               UUID         NOT NULL REFERENCES iam_users(id),

    -- Constraints
    CONSTRAINT chk_inv_txn_type
        CHECK (transaction_type IN (
            'BUY','SELL','SUBSCRIPTION','REDEMPTION',
            'CASH_IN','CASH_OUT','FEE','DIVIDEND','REVERSAL'
        )),

    CONSTRAINT chk_inv_txn_side
        CHECK (side IS NULL OR side IN ('BUY','SELL')),

    CONSTRAINT chk_inv_txn_status
        CHECK (status IN ('POSTED','REVERSED')),

    CONSTRAINT chk_inv_txn_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_txn_fees_non_negative
        CHECK (fees >= 0),

    CONSTRAINT chk_inv_txn_security_fields
        CHECK (
            transaction_type NOT IN ('BUY','SELL','SUBSCRIPTION','REDEMPTION')
            OR (instrument_id IS NOT NULL AND quantity IS NOT NULL AND price IS NOT NULL)
        ),

    CONSTRAINT chk_inv_txn_cash_only_fields
        CHECK (
            transaction_type NOT IN ('CASH_IN','CASH_OUT','FEE','DIVIDEND')
            OR (instrument_id IS NULL AND quantity IS NULL AND price IS NULL)
        ),

    CONSTRAINT chk_inv_txn_reversal
        CHECK (
            (transaction_type = 'REVERSAL' AND reverses_transaction_id IS NOT NULL)
            OR (transaction_type <> 'REVERSAL' AND reverses_transaction_id IS NULL)
        ),

    CONSTRAINT fk_inv_txn_portfolio_fund
        FOREIGN KEY (portfolio_id, fund_id)
        REFERENCES investment__portfolios (id, fund_id)
);

-- Indexes
CREATE INDEX idx_inv_txn_portfolio_date
    ON investment__portfolio_transactions (portfolio_id, business_date DESC);

CREATE INDEX idx_inv_txn_fund_date
    ON investment__portfolio_transactions (fund_id, business_date DESC);

CREATE INDEX idx_inv_txn_instrument_date
    ON investment__portfolio_transactions (instrument_id, business_date DESC)
    WHERE instrument_id IS NOT NULL;

CREATE UNIQUE INDEX uq_inv_txn_external_ref
    ON investment__portfolio_transactions (portfolio_id, external_ref)
    WHERE external_ref IS NOT NULL;

CREATE UNIQUE INDEX uq_inv_txn_decision_source
    ON investment__portfolio_transactions (source_decision_id, business_date, transaction_type)
    WHERE source_decision_id IS NOT NULL;

CREATE UNIQUE INDEX uq_inv_txn_reversal_target
    ON investment__portfolio_transactions (reverses_transaction_id)
    WHERE reverses_transaction_id IS NOT NULL;

-- Triggers
CREATE TRIGGER trg_inv_portfolio_transactions_no_update
    BEFORE UPDATE ON investment__portfolio_transactions
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();

CREATE TRIGGER trg_inv_portfolio_transactions_no_delete
    BEFORE DELETE ON investment__portfolio_transactions
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();

-- Comments
COMMENT ON TABLE  investment__portfolio_transactions IS
    'Append-only portfolio ledger. UPDATE/DELETE are rejected by mutation triggers.';
COMMENT ON COLUMN investment__portfolio_transactions.net_amount IS
    'Signed net cash impact: BUY negative, SELL positive.';
COMMENT ON COLUMN investment__portfolio_transactions.fx_rate_to_base IS
    'Required when currency != portfolio.base_currency. Enforced at app layer.';
