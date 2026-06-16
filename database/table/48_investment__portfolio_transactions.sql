-- Table: investment__portfolio_transactions
-- Source: 20260428000004_investment__create_ledger.up.sql
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
