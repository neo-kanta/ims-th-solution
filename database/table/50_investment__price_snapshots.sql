-- Table: investment__price_snapshots
-- Source: 20260428000005_investment__create_pricing_valuation.up.sql
CREATE TABLE investment__price_snapshots (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    instrument_id   UUID         NOT NULL REFERENCES investment__instruments(id) ON DELETE RESTRICT,
    business_date   DATE         NOT NULL,
    price           DECIMAL(28,8) NOT NULL,
    currency        CHAR(3)      NOT NULL,
    price_source    VARCHAR(40)  NOT NULL,
    provider_ref    VARCHAR(80),
    is_stale        BOOLEAN      NOT NULL DEFAULT false,
    stale_reason    TEXT,

    captured_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by      UUID         NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_price_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_price_positive
        CHECK (price > 0)
);
