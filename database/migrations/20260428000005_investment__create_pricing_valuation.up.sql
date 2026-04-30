-- =============================================================================
-- Investment Module — Pricing & Valuation Snapshots
-- =============================================================================
-- All snapshot tables here are append-only. Each row captures the state of
-- the market or the portfolio AS OF a business date and is reproducible from
-- its inputs (price_set_hash on valuation snapshots).
--
-- Tables:
--   investment__price_snapshots
--   investment__valuation_snapshots
--   investment__valuation_holding_lines
--   investment__nav_snapshots
--   investment__aum_snapshots
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. Price snapshots (per instrument + business_date + source)
-- ---------------------------------------------------------------------------
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

CREATE UNIQUE INDEX uq_inv_price_instrument_date_source
    ON investment__price_snapshots (instrument_id, business_date, price_source);

CREATE INDEX idx_inv_price_instrument_date_desc
    ON investment__price_snapshots (instrument_id, business_date DESC);

CREATE RULE no_update_inv_price_snapshots
    AS ON UPDATE TO investment__price_snapshots DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_price_snapshots
    AS ON DELETE TO investment__price_snapshots DO INSTEAD NOTHING;

COMMENT ON TABLE  investment__price_snapshots IS 'Append-only price observations for instruments. One row per (instrument, business_date, source).';

-- ---------------------------------------------------------------------------
-- 2. Valuation snapshots (per portfolio + business_date)
-- ---------------------------------------------------------------------------
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

CREATE UNIQUE INDEX uq_inv_valuation_portfolio_date_source
    ON investment__valuation_snapshots (portfolio_id, business_date, source);

CREATE INDEX idx_inv_valuation_portfolio_date_desc
    ON investment__valuation_snapshots (portfolio_id, business_date DESC);

CREATE RULE no_update_inv_valuation_snapshots
    AS ON UPDATE TO investment__valuation_snapshots DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_valuation_snapshots
    AS ON DELETE TO investment__valuation_snapshots DO INSTEAD NOTHING;

COMMENT ON TABLE  investment__valuation_snapshots IS 'Append-only portfolio valuation snapshot. is_indicative=true for internally computed snapshots.';
COMMENT ON COLUMN investment__valuation_snapshots.is_indicative IS 'True for internal valuations. Only PAM-sourced rows may set this false.';
COMMENT ON COLUMN investment__valuation_snapshots.price_set_hash IS 'Reproducibility hash of the price snapshot ids used.';

-- ---------------------------------------------------------------------------
-- 3. Valuation holding lines (children of a valuation snapshot)
-- ---------------------------------------------------------------------------
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

CREATE INDEX idx_inv_valuation_lines_snapshot
    ON investment__valuation_holding_lines (valuation_snapshot_id);
CREATE INDEX idx_inv_valuation_lines_instrument
    ON investment__valuation_holding_lines (instrument_id);

CREATE RULE no_update_inv_valuation_lines
    AS ON UPDATE TO investment__valuation_holding_lines DO INSTEAD NOTHING;

COMMENT ON TABLE investment__valuation_holding_lines IS 'Per-instrument detail of a valuation snapshot. Cascade-deleted only via parent (which itself is append-only at app level).';

-- ---------------------------------------------------------------------------
-- 4. NAV snapshots (only when portfolio.has_units = true)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__nav_snapshots (
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id             UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    business_date            DATE         NOT NULL,

    total_units              DECIMAL(28,8) NOT NULL,
    nav_per_unit             DECIMAL(28,12) NOT NULL,

    valuation_snapshot_id    UUID         NOT NULL REFERENCES investment__valuation_snapshots(id) ON DELETE RESTRICT,

    is_indicative            BOOLEAN      NOT NULL DEFAULT true,
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by               UUID         NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_nav_units_non_negative
        CHECK (total_units >= 0)
);

CREATE UNIQUE INDEX uq_inv_nav_portfolio_date
    ON investment__nav_snapshots (portfolio_id, business_date);

CREATE INDEX idx_inv_nav_portfolio_date_desc
    ON investment__nav_snapshots (portfolio_id, business_date DESC);

CREATE RULE no_update_inv_nav_snapshots
    AS ON UPDATE TO investment__nav_snapshots DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_nav_snapshots
    AS ON DELETE TO investment__nav_snapshots DO INSTEAD NOTHING;

COMMENT ON TABLE investment__nav_snapshots IS 'Per-portfolio NAV-per-unit snapshot. Created only when the portfolio has units issued.';

-- ---------------------------------------------------------------------------
-- 5. AUM snapshots (per portfolio AND per fund)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__aum_snapshots (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    scope_type      VARCHAR(10)  NOT NULL,
    scope_id        UUID         NOT NULL,

    business_date   DATE         NOT NULL,
    aum             DECIMAL(28,8) NOT NULL,
    valuation_ccy   CHAR(3)      NOT NULL,
    source          VARCHAR(20)  NOT NULL DEFAULT 'INTERNAL',

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by      UUID         NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_aum_scope_type
        CHECK (scope_type IN ('PORTFOLIO','FUND')),

    CONSTRAINT chk_inv_aum_ccy
        CHECK (valuation_ccy ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_aum_source
        CHECK (source IN ('INTERNAL','PAM','EXTERNAL'))
);

CREATE UNIQUE INDEX uq_inv_aum_scope_date_source
    ON investment__aum_snapshots (scope_type, scope_id, business_date, source);

CREATE INDEX idx_inv_aum_scope_date_desc
    ON investment__aum_snapshots (scope_type, scope_id, business_date DESC);

CREATE RULE no_update_inv_aum_snapshots
    AS ON UPDATE TO investment__aum_snapshots DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_aum_snapshots
    AS ON DELETE TO investment__aum_snapshots DO INSTEAD NOTHING;

COMMENT ON TABLE investment__aum_snapshots IS 'Append-only AUM snapshots. scope_type controls whether scope_id is a portfolio or fund.';
