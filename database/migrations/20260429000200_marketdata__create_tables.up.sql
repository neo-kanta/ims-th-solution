-- =============================================================================
-- Market Data Integration Module
-- =============================================================================
-- Stores provider symbol mappings, latest/historical market data snapshots,
-- and provider request logs for operational troubleshooting.
-- =============================================================================

CREATE TABLE market_symbols (
    id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol                VARCHAR(64)  NOT NULL,
    asset_type            VARCHAR(20)  NOT NULL DEFAULT 'UNKNOWN',
    name                  TEXT,
    currency              CHAR(3),
    alpha_vantage_symbol  VARCHAR(64),
    yahoo_symbol          VARCHAR(64),
    is_active             BOOLEAN      NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_market_symbols_symbol UNIQUE (symbol),
    CONSTRAINT chk_market_symbols_asset_type
        CHECK (asset_type IN ('UNKNOWN','EQUITY','ETF','MUTUAL_FUND')),
    CONSTRAINT chk_market_symbols_currency
        CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$')
);

CREATE INDEX idx_market_symbols_alpha_vantage_symbol
    ON market_symbols (alpha_vantage_symbol)
    WHERE alpha_vantage_symbol IS NOT NULL;

CREATE INDEX idx_market_symbols_yahoo_symbol
    ON market_symbols (yahoo_symbol)
    WHERE yahoo_symbol IS NOT NULL;

COMMENT ON TABLE market_symbols IS 'Canonical market data symbols and provider-specific ticker mappings.';
COMMENT ON COLUMN market_symbols.yahoo_symbol IS 'Yahoo Finance ticker mapping. Yahoo Finance is an unofficial fallback source only.';

CREATE TABLE market_data_snapshots (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol_id         UUID          NOT NULL REFERENCES market_symbols(id) ON DELETE RESTRICT,
    provider_name     VARCHAR(40)   NOT NULL,
    data_type         VARCHAR(10)   NOT NULL,
    price_date        DATE          NOT NULL,
    as_of             TIMESTAMPTZ   NOT NULL,

    price             DECIMAL(28,8) NOT NULL,
    open_price        DECIMAL(28,8),
    high_price        DECIMAL(28,8),
    low_price         DECIMAL(28,8),
    close_price       DECIMAL(28,8) NOT NULL,
    adjusted_close    DECIMAL(28,8),
    previous_close    DECIMAL(28,8),
    change_amount     DECIMAL(28,8),
    change_percent    DECIMAL(18,8),
    volume            BIGINT,
    currency          CHAR(3),
    raw_payload       JSONB         NOT NULL DEFAULT '{}'::jsonb,

    captured_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_market_data_snapshots_provider_name
        CHECK (provider_name <> ''),
    CONSTRAINT chk_market_data_snapshots_type
        CHECK (data_type IN ('QUOTE','DAILY')),
    CONSTRAINT chk_market_data_snapshots_positive_price
        CHECK (price > 0 AND close_price > 0),
    CONSTRAINT chk_market_data_snapshots_currency
        CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$')
);

CREATE UNIQUE INDEX uq_market_data_symbol_provider_type_date
    ON market_data_snapshots (symbol_id, provider_name, data_type, price_date);

CREATE INDEX idx_market_data_symbol_type_date_desc
    ON market_data_snapshots (symbol_id, data_type, price_date DESC);

CREATE INDEX idx_market_data_provider_captured_desc
    ON market_data_snapshots (provider_name, captured_at DESC);

COMMENT ON TABLE market_data_snapshots IS 'Latest quote and historical daily snapshots received from market data providers.';

CREATE TABLE provider_requests_log (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_name   VARCHAR(40)  NOT NULL,
    symbol          VARCHAR(64)  NOT NULL,
    operation       VARCHAR(20)  NOT NULL,
    status          VARCHAR(20)  NOT NULL,
    status_code     INTEGER,
    error_code      VARCHAR(80),
    error_message   TEXT,
    duration_ms     INTEGER      NOT NULL DEFAULT 0,
    requested_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_provider_requests_log_provider_name
        CHECK (provider_name <> ''),
    CONSTRAINT chk_provider_requests_log_operation
        CHECK (operation IN ('quote','history')),
    CONSTRAINT chk_provider_requests_log_status
        CHECK (status IN ('success','error'))
);

CREATE INDEX idx_provider_requests_log_provider_requested
    ON provider_requests_log (provider_name, requested_at DESC);

CREATE INDEX idx_provider_requests_log_symbol_requested
    ON provider_requests_log (symbol, requested_at DESC);

COMMENT ON TABLE provider_requests_log IS 'Operational request log for market data provider calls.';
