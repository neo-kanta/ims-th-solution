-- Table: market_symbols
-- Source: 20260429000200_marketdata__create_tables.up.sql
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
