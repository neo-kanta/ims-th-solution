-- Table: investment__cash_balances
-- Source: 20260428000004_investment__create_ledger.up.sql
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
