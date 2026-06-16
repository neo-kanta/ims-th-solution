-- Table: investment__valuation_snapshots
-- Source: 20260428000005_investment__create_pricing_valuation.up.sql
CREATE TABLE investment__valuation_snapshots (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id       UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    business_date      DATE         NOT NULL,
    valuation_ccy      CHAR(3)      NOT NULL,

    market_value       DECIMAL(28,8) NOT NULL,
    cost_basis         DECIMAL(28,8) NOT NULL,
    unrealised_pnl     DECIMAL(28,8) NOT NULL,
    realised_pnl       DECIMAL(28,8) NOT NULL DEFAULT 0,
    roi                DECIMAL(18,8),
    aum                DECIMAL(28,8) NOT NULL,
    cash_balance       DECIMAL(28,8) NOT NULL,

    price_set_hash     VARCHAR(64)  NOT NULL,
    has_stale_inputs   BOOLEAN      NOT NULL DEFAULT false,
    is_indicative      BOOLEAN      NOT NULL DEFAULT true,
    source             VARCHAR(20)  NOT NULL DEFAULT 'INTERNAL',

    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by         UUID         NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_valuation_ccy
        CHECK (valuation_ccy ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_valuation_source
        CHECK (source IN ('INTERNAL','PAM','EXTERNAL'))
);
