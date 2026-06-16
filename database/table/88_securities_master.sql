-- Table: securities_master
-- Source: 20260526000002_referencedata__create_securities.up.sql
CREATE TABLE securities_master (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    ims_symbol          VARCHAR(128) NOT NULL,
    display_symbol      VARCHAR(128) NOT NULL,
    primary_identifier  VARCHAR(128),
    name                TEXT         NOT NULL,
    asset_type          VARCHAR(30)  NOT NULL,
    currency            CHAR(3),
    country_code        CHAR(2),
    exchange_mic        VARCHAR(16),
    isin                VARCHAR(12),
    cusip               VARCHAR(16),
    figi                VARCHAR(32),
    status              VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE',
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_securities_master_ims_symbol UNIQUE (ims_symbol),
    CONSTRAINT chk_securities_master_asset_type
        CHECK (asset_type IN (
            'UNKNOWN','EQUITY','ETF','MUTUAL_FUND','BOND','FX','INDEX','CASH','DERIVATIVE','OTHER'
        )),
    CONSTRAINT chk_securities_master_status
        CHECK (status IN ('ACTIVE','INACTIVE','SUSPENDED')),
    CONSTRAINT chk_securities_master_currency
        CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_securities_master_country
        CHECK (country_code IS NULL OR country_code ~ '^[A-Z]{2}$'),
    CONSTRAINT chk_securities_master_isin_len
        CHECK (isin IS NULL OR char_length(isin) = 12)
);
