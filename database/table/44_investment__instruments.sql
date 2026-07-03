-- Table: investment__instruments
-- Source: 20260428000003_investment__create_instruments.up.sql
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
    provider_symbol_alpha_vantage VARCHAR(64),
    provider_symbol_yahoo         VARCHAR(64),

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
