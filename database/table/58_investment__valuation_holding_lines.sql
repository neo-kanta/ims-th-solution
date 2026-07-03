-- Table: investment__valuation_holding_lines
-- Source: 20260428000005_investment__create_pricing_valuation.up.sql
CREATE TABLE investment__valuation_holding_lines (
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    valuation_snapshot_id    UUID         NOT NULL REFERENCES investment__valuation_snapshots(id) ON DELETE CASCADE,
    instrument_id            UUID         NOT NULL REFERENCES investment__instruments(id) ON DELETE RESTRICT,
    price_snapshot_id        UUID         REFERENCES investment__price_snapshots(id) ON DELETE RESTRICT,

    quantity                 DECIMAL(28,8) NOT NULL,
    price_in_quote_ccy       DECIMAL(28,8) NOT NULL,
    quote_currency           CHAR(3)       NOT NULL,
    fx_rate_to_valuation_ccy DECIMAL(28,12) NOT NULL,

    market_value             DECIMAL(28,8) NOT NULL,
    cost_basis               DECIMAL(28,8) NOT NULL,
    unrealised_pnl           DECIMAL(28,8) NOT NULL,
    is_stale                 BOOLEAN       NOT NULL DEFAULT false,

    created_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_inv_valuation_line_quote_ccy
        CHECK (quote_currency ~ '^[A-Z]{3}$')
);
