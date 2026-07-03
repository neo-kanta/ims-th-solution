-- Table: market_data_snapshots
-- Source: 20260429000200_marketdata__create_tables.up.sql
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
