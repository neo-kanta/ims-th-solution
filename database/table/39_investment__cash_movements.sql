-- Table: investment__cash_movements
-- Source: 20260428000004_investment__create_ledger.up.sql
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
