-- =============================================================================
-- Investment Module — Instrument / Security Master + Identifiers
-- =============================================================================
-- One row per tradable security. `attributes JSONB` holds asset-class-specific
-- fields (coupon_rate, maturity, strike, multiplier, etc.) so adding bond /
-- derivative support later does not require a schema migration.
--
-- Identifiers (ISIN, CUSIP, SEDOL, BBG, RIC, vendor IDs) live in a separate
-- many-to-one table to support multiple identifiers per instrument.
-- =============================================================================

CREATE TABLE investment__instruments (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    primary_ticker      VARCHAR(40)  NOT NULL,
    name                VARCHAR(255) NOT NULL,

    asset_class_id      UUID         NOT NULL REFERENCES investment__asset_classes(id) ON DELETE RESTRICT,
    asset_subtype_id    UUID         NOT NULL REFERENCES investment__asset_subtypes(id) ON DELETE RESTRICT,

    currency            CHAR(3)      NOT NULL,
    country_id          UUID         NOT NULL REFERENCES investment__countries(id) ON DELETE RESTRICT,
    region_id           UUID         REFERENCES investment__regions(id) ON DELETE RESTRICT,

    primary_exchange    VARCHAR(20),

    sector_id           UUID         REFERENCES investment__sectors(id) ON DELETE RESTRICT,
    fund_category_id    UUID         REFERENCES investment__fund_categories(id) ON DELETE RESTRICT,

    lot_size            INTEGER      NOT NULL DEFAULT 1,
    tick_size           DECIMAL(20,8),

    is_tradable         BOOLEAN      NOT NULL DEFAULT true,
    status              VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE',

    attributes          JSONB        NOT NULL DEFAULT '{}'::jsonb,

    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by          UUID         REFERENCES iam_users(id),
    updated_by          UUID         REFERENCES iam_users(id),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT chk_inv_instruments_status
        CHECK (status IN ('ACTIVE','SUSPENDED','DELISTED')),

    CONSTRAINT chk_inv_instruments_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_instruments_lot_positive
        CHECK (lot_size >= 1),

    CONSTRAINT chk_inv_instruments_tick_positive
        CHECK (tick_size IS NULL OR tick_size > 0)
);

CREATE UNIQUE INDEX uq_inv_instruments_ticker_exchange_alive
    ON investment__instruments (primary_ticker, COALESCE(primary_exchange, ''))
    WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_instruments_class     ON investment__instruments (asset_class_id, asset_subtype_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_inv_instruments_sector    ON investment__instruments (sector_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_inv_instruments_country   ON investment__instruments (country_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_inv_instruments_status    ON investment__instruments (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_inv_instruments_attrs     ON investment__instruments USING GIN (attributes);

CREATE TRIGGER trg_inv_instruments_updated_at
    BEFORE UPDATE ON investment__instruments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE  investment__instruments IS 'Security / instrument master. Asset-class-specific fields go in attributes JSONB.';
COMMENT ON COLUMN investment__instruments.attributes IS 'Future-friendly bag for bond (coupon, maturity), derivative (strike, multiplier), MMF (min subscription) fields.';

-- ---------------------------------------------------------------------------
-- Instrument identifiers (ISIN, CUSIP, SEDOL, BBG, RIC, provider IDs)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__instrument_identifiers (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    instrument_id   UUID         NOT NULL REFERENCES investment__instruments(id) ON DELETE CASCADE,
    id_type         VARCHAR(20)  NOT NULL,
    id_value        VARCHAR(60)  NOT NULL,
    provider_code   VARCHAR(40),
    is_primary      BOOLEAN      NOT NULL DEFAULT false,

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by      UUID         REFERENCES iam_users(id),
    updated_by      UUID         REFERENCES iam_users(id),

    CONSTRAINT chk_inv_instrument_identifiers_type
        CHECK (id_type IN ('ISIN','CUSIP','SEDOL','BBG','RIC','MORNINGSTAR','PROVIDER'))
);

CREATE UNIQUE INDEX uq_inv_instrument_identifiers_type_value
    ON investment__instrument_identifiers (id_type, id_value);

CREATE UNIQUE INDEX uq_inv_instrument_identifiers_provider
    ON investment__instrument_identifiers (instrument_id, id_type, COALESCE(provider_code, ''));

CREATE INDEX idx_inv_instrument_identifiers_instrument
    ON investment__instrument_identifiers (instrument_id);

CREATE TRIGGER trg_inv_instrument_identifiers_updated_at
    BEFORE UPDATE ON investment__instrument_identifiers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__instrument_identifiers IS 'External identifier mappings for an instrument. ISIN/CUSIP/etc.';
