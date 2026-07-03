-- =============================================================================
-- Reference Data — canonical securities, provider mappings, unmapped candidates
-- =============================================================================
-- Establishes IMS canonical security identity (securities_master), a generic
-- many-to-one provider-symbol mapping table (security_provider_mappings), and
-- an unmapped-candidate review workflow (reference_data_unmapped_security_candidates).
--
-- Legacy compatibility:
--   * market_symbols stays in place.
--   * market_symbols.security_id is added as a nullable FK so market_data can
--     start linking snapshots to canonical securities without a hard cutover.
--   * Existing market_symbols rows are backfilled into securities_master, and
--     non-empty alpha_vantage_symbol / yahoo_symbol values are projected into
--     security_provider_mappings.
-- =============================================================================

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

CREATE UNIQUE INDEX uq_securities_master_isin
    ON securities_master (isin)
    WHERE isin IS NOT NULL;

CREATE INDEX idx_securities_master_display_symbol ON securities_master (display_symbol);
CREATE INDEX idx_securities_master_asset_type    ON securities_master (asset_type);
CREATE INDEX idx_securities_master_exchange_mic  ON securities_master (exchange_mic);
CREATE INDEX idx_securities_master_status        ON securities_master (status);

COMMENT ON TABLE securities_master IS
    'Canonical IMS security identity. ims_symbol is the stable internal ID; provider symbols live in security_provider_mappings.';
COMMENT ON COLUMN securities_master.ims_symbol IS
    'Stable internal identity. Examples: TH_EQ_XBKK_KBANK, US_BOND_US912828Z948, FX_USDTHB.';
COMMENT ON COLUMN securities_master.display_symbol IS
    'Human-facing symbol shown in UIs. Not unique; multiple markets may share a display string.';

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

CREATE UNIQUE INDEX uq_security_provider_mappings_provider_symbol
    ON security_provider_mappings (provider_code, provider_symbol);
CREATE INDEX idx_security_provider_mappings_security ON security_provider_mappings (security_id);
CREATE INDEX idx_security_provider_mappings_provider ON security_provider_mappings (provider_code);
CREATE INDEX idx_security_provider_mappings_status   ON security_provider_mappings (mapping_status);

COMMENT ON TABLE security_provider_mappings IS
    'Many-to-one mapping from (provider_code, provider_symbol) to a canonical IMS security.';
COMMENT ON COLUMN security_provider_mappings.is_primary IS
    'Marks the canonical provider symbol the system should prefer when multiple mappings exist for the same provider.';

CREATE TABLE reference_data_unmapped_security_candidates (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id               UUID,
    provider_code          VARCHAR(40)  NOT NULL,
    provider_symbol        VARCHAR(128) NOT NULL,
    provider_name          TEXT,
    provider_asset_type    VARCHAR(40),
    provider_exchange      VARCHAR(64),
    provider_currency      CHAR(3),
    isin                   VARCHAR(12),
    raw_payload            JSONB        NOT NULL DEFAULT '{}'::jsonb,
    candidate_status       VARCHAR(30)  NOT NULL DEFAULT 'REVIEW_REQUIRED',
    suggested_security_id  UUID         REFERENCES securities_master(id) ON DELETE SET NULL,
    confidence_score       NUMERIC(5,2),
    rejected_reason        TEXT,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    resolved_at            TIMESTAMPTZ,

    CONSTRAINT chk_unmapped_candidates_status
        CHECK (candidate_status IN ('REVIEW_REQUIRED','MAPPED','REJECTED','DUPLICATE','CONFLICTED')),
    CONSTRAINT chk_unmapped_candidates_provider_code
        CHECK (provider_code <> ''),
    CONSTRAINT chk_unmapped_candidates_provider_symbol
        CHECK (provider_symbol <> ''),
    CONSTRAINT chk_unmapped_candidates_currency
        CHECK (provider_currency IS NULL OR provider_currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_unmapped_candidates_isin_len
        CHECK (isin IS NULL OR char_length(isin) = 12)
);

CREATE UNIQUE INDEX uq_unmapped_candidates_open
    ON reference_data_unmapped_security_candidates (provider_code, provider_symbol)
    WHERE candidate_status = 'REVIEW_REQUIRED';

CREATE INDEX idx_unmapped_candidates_batch     ON reference_data_unmapped_security_candidates (batch_id);
CREATE INDEX idx_unmapped_candidates_status    ON reference_data_unmapped_security_candidates (candidate_status);
CREATE INDEX idx_unmapped_candidates_suggested ON reference_data_unmapped_security_candidates (suggested_security_id);

COMMENT ON TABLE reference_data_unmapped_security_candidates IS
    'Review queue for provider symbols that did not resolve to a canonical IMS security.';

-- =============================================================================
-- Legacy compatibility — link market_symbols to canonical securities.
-- =============================================================================
ALTER TABLE market_symbols
    ADD COLUMN security_id UUID NULL REFERENCES securities_master(id) ON DELETE SET NULL;

CREATE INDEX idx_market_symbols_security_id
    ON market_symbols (security_id)
    WHERE security_id IS NOT NULL;

COMMENT ON COLUMN market_symbols.security_id IS
    'Pointer to canonical IMS security in securities_master. Nullable during the legacy transition.';

-- =============================================================================
-- Backfill: insert one securities_master per existing market_symbols row,
-- project alpha_vantage_symbol / yahoo_symbol into security_provider_mappings.
-- Re-running this block is safe; ON CONFLICT guards make it idempotent.
-- =============================================================================
INSERT INTO securities_master (id, ims_symbol, display_symbol, name, asset_type, currency, status)
SELECT
    gen_random_uuid(),
    'SEC_' || UPPER(REGEXP_REPLACE(ms.symbol, '[^A-Za-z0-9]+', '_', 'g')),
    ms.symbol,
    COALESCE(NULLIF(ms.name, ''), ms.symbol),
    CASE ms.asset_type
        WHEN 'EQUITY'      THEN 'EQUITY'
        WHEN 'ETF'         THEN 'ETF'
        WHEN 'MUTUAL_FUND' THEN 'MUTUAL_FUND'
        ELSE                    'UNKNOWN'
    END,
    NULLIF(UPPER(ms.currency), ''),
    'ACTIVE'
  FROM market_symbols ms
 WHERE NOT EXISTS (
        SELECT 1
          FROM securities_master sm
         WHERE sm.ims_symbol = 'SEC_' || UPPER(REGEXP_REPLACE(ms.symbol, '[^A-Za-z0-9]+', '_', 'g'))
       )
ON CONFLICT (ims_symbol) DO NOTHING;

UPDATE market_symbols ms
   SET security_id = sm.id
  FROM securities_master sm
 WHERE ms.security_id IS NULL
   AND sm.ims_symbol = 'SEC_' || UPPER(REGEXP_REPLACE(ms.symbol, '[^A-Za-z0-9]+', '_', 'g'));

-- Project legacy alpha_vantage_symbol values into provider mappings.
INSERT INTO security_provider_mappings (
    security_id, provider_code, provider_symbol, provider_currency, mapping_status, is_primary, priority
)
SELECT
    ms.security_id,
    'alpha_vantage',
    ms.alpha_vantage_symbol,
    NULLIF(UPPER(ms.currency), ''),
    'ACTIVE',
    (ms.alpha_vantage_symbol = ms.symbol),
    50
  FROM market_symbols ms
 WHERE ms.security_id IS NOT NULL
   AND COALESCE(NULLIF(ms.alpha_vantage_symbol, ''), NULL) IS NOT NULL
ON CONFLICT (provider_code, provider_symbol) DO NOTHING;

-- Project legacy yahoo_symbol values into provider mappings.
INSERT INTO security_provider_mappings (
    security_id, provider_code, provider_symbol, provider_currency, mapping_status, is_primary, priority
)
SELECT
    ms.security_id,
    'yahoo',
    ms.yahoo_symbol,
    NULLIF(UPPER(ms.currency), ''),
    'ACTIVE',
    (ms.yahoo_symbol = ms.symbol),
    100
  FROM market_symbols ms
 WHERE ms.security_id IS NOT NULL
   AND COALESCE(NULLIF(ms.yahoo_symbol, ''), NULL) IS NOT NULL
ON CONFLICT (provider_code, provider_symbol) DO NOTHING;
