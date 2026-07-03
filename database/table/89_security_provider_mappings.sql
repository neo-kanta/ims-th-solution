-- Table: security_provider_mappings
-- Source: 20260526000002_referencedata__create_securities.up.sql
CREATE TABLE security_provider_mappings (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    security_id         UUID          NOT NULL REFERENCES securities_master(id) ON DELETE CASCADE,
    provider_code       VARCHAR(40)   NOT NULL,
    provider_symbol     VARCHAR(128)  NOT NULL,
    provider_exchange   VARCHAR(64),
    provider_asset_type VARCHAR(40),
    provider_currency   CHAR(3),
    priority            INTEGER       NOT NULL DEFAULT 100,
    confidence_score    NUMERIC(5,2)  NOT NULL DEFAULT 100.00,
    mapping_status      VARCHAR(30)   NOT NULL DEFAULT 'ACTIVE',
    is_primary          BOOLEAN       NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_security_provider_mappings_provider_code
        CHECK (provider_code <> ''),
    CONSTRAINT chk_security_provider_mappings_provider_symbol
        CHECK (provider_symbol <> ''),
    CONSTRAINT chk_security_provider_mappings_status
        CHECK (mapping_status IN ('ACTIVE','INACTIVE','UNMAPPED','CONFLICTED','REVIEW_REQUIRED')),
    CONSTRAINT chk_security_provider_mappings_currency
        CHECK (provider_currency IS NULL OR provider_currency ~ '^[A-Z]{3}$')
);
