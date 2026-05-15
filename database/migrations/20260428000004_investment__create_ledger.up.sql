-- =============================================================================
-- Investment Module — Portfolio Ledger
-- =============================================================================
-- Append-only transaction ledger + position projection + cash ledger.
--
-- Append-only enforcement (defence in depth):
--   * Application-layer guard (commands never UPDATE/DELETE these rows)
--   * Postgres RULE rewrites UPDATE/DELETE to NOTHING
--
-- Tables created here:
--   investment__portfolio_transactions   (immutable)
--   investment__portfolio_positions      (mutable projection, version-locked)
--   investment__cash_movements           (immutable)
--   investment__cash_balances            (mutable projection, version-locked)
--
-- Decimal scales:
--   quantity / price / money:  DECIMAL(28,8)
--   FX:                        DECIMAL(28,12)
--   ROI fraction (later):      DECIMAL(18,8)
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. Portfolio transactions (immutable ledger)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__portfolio_transactions (
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id             UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    fund_id                  UUID         NOT NULL REFERENCES investment__funds(id)      ON DELETE RESTRICT,
    instrument_id            UUID         REFERENCES investment__instruments(id)         ON DELETE RESTRICT,

    transaction_type         VARCHAR(20)  NOT NULL,
    side                     VARCHAR(10),

    quantity                 DECIMAL(28,8),
    price                    DECIMAL(28,8),
    currency                 CHAR(3)      NOT NULL,
    gross_amount             DECIMAL(28,8) NOT NULL,
    fees                     DECIMAL(28,8) NOT NULL DEFAULT 0,
    net_amount               DECIMAL(28,8) NOT NULL,
    fx_rate_to_base          DECIMAL(28,12),

    business_date            DATE         NOT NULL,
    settlement_date          DATE,

    source_decision_id       UUID,
    source_execution_id      UUID,
    reverses_transaction_id  UUID         REFERENCES investment__portfolio_transactions(id) ON DELETE RESTRICT,

    external_ref             VARCHAR(80),
    reason                   TEXT,

    status                   VARCHAR(20)  NOT NULL DEFAULT 'POSTED',

    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by               UUID         NOT NULL REFERENCES iam_users(id),

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

    -- security trades require quantity, price, instrument
    CONSTRAINT chk_inv_txn_security_fields
        CHECK (
            transaction_type NOT IN ('BUY','SELL','SUBSCRIPTION','REDEMPTION')
            OR (instrument_id IS NOT NULL AND quantity IS NOT NULL AND price IS NOT NULL)
        ),

    -- pure cash movements must NOT carry instrument/quantity
    CONSTRAINT chk_inv_txn_cash_only_fields
        CHECK (
            transaction_type NOT IN ('CASH_IN','CASH_OUT','FEE','DIVIDEND')
            OR (instrument_id IS NULL AND quantity IS NULL AND price IS NULL)
        ),

    -- reversal must reference an original
    CONSTRAINT chk_inv_txn_reversal
        CHECK (
            (transaction_type = 'REVERSAL' AND reverses_transaction_id IS NOT NULL)
            OR (transaction_type <> 'REVERSAL' AND reverses_transaction_id IS NULL)
        )
);

-- Hot-path indexes
CREATE INDEX idx_inv_txn_portfolio_date
    ON investment__portfolio_transactions (portfolio_id, business_date DESC);
CREATE INDEX idx_inv_txn_fund_date
    ON investment__portfolio_transactions (fund_id, business_date DESC);
CREATE INDEX idx_inv_txn_instrument_date
    ON investment__portfolio_transactions (instrument_id, business_date DESC)
    WHERE instrument_id IS NOT NULL;

-- Idempotency: prevent duplicate broker confirmations.
CREATE UNIQUE INDEX uq_inv_txn_external_ref
    ON investment__portfolio_transactions (portfolio_id, external_ref)
    WHERE external_ref IS NOT NULL;

-- Idempotency: one ledger post per (decision, business_date, type).
CREATE UNIQUE INDEX uq_inv_txn_decision_source
    ON investment__portfolio_transactions (source_decision_id, business_date, transaction_type)
    WHERE source_decision_id IS NOT NULL;

-- A transaction may be reversed at most once.
CREATE UNIQUE INDEX uq_inv_txn_reversal_target
    ON investment__portfolio_transactions (reverses_transaction_id)
    WHERE reverses_transaction_id IS NOT NULL;

-- Append-only enforcement
CREATE RULE no_update_inv_portfolio_transactions
    AS ON UPDATE TO investment__portfolio_transactions DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_portfolio_transactions
    AS ON DELETE TO investment__portfolio_transactions DO INSTEAD NOTHING;

COMMENT ON TABLE  investment__portfolio_transactions IS 'Append-only portfolio ledger. UPDATE/DELETE rewritten to NOTHING by RULE.';
COMMENT ON COLUMN investment__portfolio_transactions.net_amount IS 'Signed net cash impact: BUY negative, SELL positive.';
COMMENT ON COLUMN investment__portfolio_transactions.fx_rate_to_base IS 'Required when currency != portfolio.base_currency. Enforced at app layer.';

-- ---------------------------------------------------------------------------
-- 2. Portfolio position projection
-- ---------------------------------------------------------------------------
CREATE TABLE investment__portfolio_positions (
    id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id          UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    instrument_id         UUID         NOT NULL REFERENCES investment__instruments(id) ON DELETE RESTRICT,

    quantity              DECIMAL(28,8) NOT NULL DEFAULT 0,
    average_cost          DECIMAL(28,8) NOT NULL DEFAULT 0,
    cost_basis            DECIMAL(28,8) NOT NULL DEFAULT 0,

    last_transaction_id   UUID         REFERENCES investment__portfolio_transactions(id) ON DELETE RESTRICT,
    last_business_date    DATE,

    version               INTEGER      NOT NULL DEFAULT 1,
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_inv_position_portfolio_instrument
        UNIQUE (portfolio_id, instrument_id),

    CONSTRAINT chk_inv_position_quantity_non_negative
        CHECK (quantity >= 0),

    CONSTRAINT chk_inv_position_avg_cost_non_negative
        CHECK (average_cost >= 0),

    CONSTRAINT chk_inv_position_version_positive
        CHECK (version >= 1)
);

CREATE INDEX idx_inv_position_portfolio
    ON investment__portfolio_positions (portfolio_id);
CREATE INDEX idx_inv_position_instrument
    ON investment__portfolio_positions (instrument_id);

CREATE TRIGGER trg_inv_position_updated_at
    BEFORE UPDATE ON investment__portfolio_positions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE  investment__portfolio_positions IS 'Current projection of holdings derived from the immutable ledger.';
COMMENT ON COLUMN investment__portfolio_positions.average_cost IS 'Average cost per unit in portfolio base currency.';

-- ---------------------------------------------------------------------------
-- 3. Cash movements (immutable)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__cash_movements (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id    UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    currency        CHAR(3)      NOT NULL,
    amount          DECIMAL(28,8) NOT NULL,                       -- signed
    business_date   DATE         NOT NULL,

    transaction_id  UUID         REFERENCES investment__portfolio_transactions(id) ON DELETE RESTRICT,
    movement_type   VARCHAR(20)  NOT NULL,

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by      UUID         NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_cash_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_cash_movement_type
        CHECK (movement_type IN (
            'OPENING_BALANCE','TRADE','SUBSCRIPTION','REDEMPTION',
            'CASH_IN','CASH_OUT','FEE','DIVIDEND','REVERSAL','FX_CONV'
        ))
);

CREATE INDEX idx_inv_cash_portfolio_date
    ON investment__cash_movements (portfolio_id, business_date DESC);
CREATE INDEX idx_inv_cash_txn
    ON investment__cash_movements (transaction_id)
    WHERE transaction_id IS NOT NULL;

CREATE RULE no_update_inv_cash_movements
    AS ON UPDATE TO investment__cash_movements DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_cash_movements
    AS ON DELETE TO investment__cash_movements DO INSTEAD NOTHING;

COMMENT ON TABLE investment__cash_movements IS 'Append-only cash movement ledger per portfolio + currency.';

-- ---------------------------------------------------------------------------
-- 4. Cash balances (mutable projection)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__cash_balances (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id        UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    currency            CHAR(3)      NOT NULL,

    balance             DECIMAL(28,8) NOT NULL DEFAULT 0,
    last_movement_id    UUID         REFERENCES investment__cash_movements(id) ON DELETE RESTRICT,
    last_business_date  DATE,

    version             INTEGER      NOT NULL DEFAULT 1,
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_inv_cash_balance_portfolio_currency
        UNIQUE (portfolio_id, currency),

    CONSTRAINT chk_inv_cash_balance_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_cash_balance_version_positive
        CHECK (version >= 1)
);

CREATE INDEX idx_inv_cash_balance_portfolio
    ON investment__cash_balances (portfolio_id);

CREATE TRIGGER trg_inv_cash_balance_updated_at
    BEFORE UPDATE ON investment__cash_balances
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__cash_balances IS 'Per (portfolio, currency) cash balance projection. Negative balances are allowed (overdraft scenarios) but flagged at the app layer.';
