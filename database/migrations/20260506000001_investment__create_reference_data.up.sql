-- =============================================================================
-- Investment Module - Reference Data Tables (Phase 0)
-- =============================================================================
-- These seven tables provide the static taxonomy required by Phase 1 entities
-- (instruments, funds, portfolios). Phase 0 owns only the tables and the
-- accompanying seed data; runtime entities are introduced in Phase 1.
--
-- Conventions:
--   - All tables use code as a stable, human-readable, UPPER_SNAKE primary key.
--   - Each table is module-prefixed (investment__*) to keep the global namespace
--     unambiguous, matching investment__process_steps and the broader module
--     convention.
--   - All tables carry created_at / updated_at timestamps with the standard
--     set_updated_at trigger, even when current callers do not mutate rows; the
--     fields are required for audit consistency once Phase 1 admin UIs land.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. Asset Classes
-- ---------------------------------------------------------------------------
CREATE TABLE investment__asset_classes (
    code            VARCHAR(40)  PRIMARY KEY,
    name            VARCHAR(120) NOT NULL,
    description     TEXT,
    sort_order      SMALLINT     NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_investment_asset_classes_code
        CHECK (code = UPPER(code) AND code !~ '\s')
);

CREATE INDEX idx_investment_asset_classes_active ON investment__asset_classes (is_active, sort_order);

CREATE TRIGGER trg_investment_asset_classes_updated_at
    BEFORE UPDATE ON investment__asset_classes
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__asset_classes IS 'Top-level asset class taxonomy (EQUITY, FIXED_INCOME, FUND, ETF, CASH, ALTERNATIVE, DERIVATIVE).';

-- ---------------------------------------------------------------------------
-- 2. Asset Subtypes
-- ---------------------------------------------------------------------------
CREATE TABLE investment__asset_subtypes (
    code            VARCHAR(40)  PRIMARY KEY,
    asset_class     VARCHAR(40)  NOT NULL REFERENCES investment__asset_classes(code) ON UPDATE RESTRICT ON DELETE RESTRICT,
    name            VARCHAR(120) NOT NULL,
    description     TEXT,
    sort_order      SMALLINT     NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_investment_asset_subtypes_code
        CHECK (code = UPPER(code) AND code !~ '\s')
);

CREATE INDEX idx_investment_asset_subtypes_class ON investment__asset_subtypes (asset_class, sort_order);
CREATE INDEX idx_investment_asset_subtypes_active ON investment__asset_subtypes (is_active);

CREATE TRIGGER trg_investment_asset_subtypes_updated_at
    BEFORE UPDATE ON investment__asset_subtypes
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__asset_subtypes IS 'Asset subtype taxonomy keyed under an asset class (e.g., COMMON_STOCK under EQUITY).';

-- ---------------------------------------------------------------------------
-- 3. Regions
-- ---------------------------------------------------------------------------
CREATE TABLE investment__regions (
    code            VARCHAR(40)  PRIMARY KEY,
    name            VARCHAR(120) NOT NULL,
    description     TEXT,
    sort_order      SMALLINT     NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_investment_regions_code
        CHECK (code = UPPER(code) AND code !~ '\s')
);

CREATE INDEX idx_investment_regions_active ON investment__regions (is_active, sort_order);

CREATE TRIGGER trg_investment_regions_updated_at
    BEFORE UPDATE ON investment__regions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__regions IS 'Coarse geographic groupings used for instrument classification (APAC, EMEA, AMER, GLOBAL, THAILAND, EMERGING, DEVELOPED).';

-- ---------------------------------------------------------------------------
-- 4. Countries (ISO-3166-1 alpha-2)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__countries (
    code            CHAR(2)      PRIMARY KEY,
    alpha3          CHAR(3)      NOT NULL,
    numeric_code    CHAR(3)      NOT NULL,
    name            VARCHAR(120) NOT NULL,
    region_code     VARCHAR(40)  NOT NULL REFERENCES investment__regions(code) ON UPDATE RESTRICT ON DELETE RESTRICT,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_investment_countries_code
        CHECK (code = UPPER(code) AND alpha3 = UPPER(alpha3) AND code !~ '\s'),
    CONSTRAINT uq_investment_countries_alpha3 UNIQUE (alpha3),
    CONSTRAINT uq_investment_countries_numeric UNIQUE (numeric_code)
);

CREATE INDEX idx_investment_countries_region ON investment__countries (region_code);
CREATE INDEX idx_investment_countries_active ON investment__countries (is_active);

CREATE TRIGGER trg_investment_countries_updated_at
    BEFORE UPDATE ON investment__countries
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__countries IS 'ISO-3166-1 country list with alpha-2 PK, alpha-3 and numeric secondary keys, plus region grouping.';

-- ---------------------------------------------------------------------------
-- 5. Sectors (GICS levels 1 + 2)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__sectors (
    code            VARCHAR(40)  PRIMARY KEY,
    parent_code     VARCHAR(40)  REFERENCES investment__sectors(code) ON UPDATE RESTRICT ON DELETE RESTRICT,
    level           SMALLINT     NOT NULL,
    gics_code       VARCHAR(20),
    name            VARCHAR(120) NOT NULL,
    description     TEXT,
    sort_order      SMALLINT     NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_investment_sectors_code
        CHECK (code = UPPER(code) AND code !~ '\s'),
    CONSTRAINT chk_investment_sectors_level
        CHECK (level BETWEEN 1 AND 4),
    CONSTRAINT chk_investment_sectors_parent
        CHECK (
            (level = 1 AND parent_code IS NULL)
            OR (level > 1 AND parent_code IS NOT NULL)
        )
);

CREATE INDEX idx_investment_sectors_parent ON investment__sectors (parent_code);
CREATE INDEX idx_investment_sectors_level ON investment__sectors (level, sort_order);
CREATE INDEX idx_investment_sectors_active ON investment__sectors (is_active);

CREATE TRIGGER trg_investment_sectors_updated_at
    BEFORE UPDATE ON investment__sectors
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__sectors IS 'Hierarchical sector taxonomy. PoC seeds GICS Level 1 (sectors) and Level 2 (industry groups); Level 3 and 4 are reserved for a future phase.';
COMMENT ON COLUMN investment__sectors.gics_code IS 'Optional GICS numeric code (string for leading-zero preservation). Labels are project-authored to avoid MSCI proprietary text.';

-- ---------------------------------------------------------------------------
-- 6. Fund Categories
-- ---------------------------------------------------------------------------
CREATE TABLE investment__fund_categories (
    code            VARCHAR(40)  PRIMARY KEY,
    name            VARCHAR(120) NOT NULL,
    description     TEXT,
    sort_order      SMALLINT     NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_investment_fund_categories_code
        CHECK (code = UPPER(code) AND code !~ '\s')
);

CREATE INDEX idx_investment_fund_categories_active ON investment__fund_categories (is_active, sort_order);

CREATE TRIGGER trg_investment_fund_categories_updated_at
    BEFORE UPDATE ON investment__fund_categories
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__fund_categories IS 'Fund category taxonomy used at fund-creation time (EQUITY_FUND, BOND_FUND, MIXED_FUND, MMF_FUND, INDEX_FUND).';

-- ---------------------------------------------------------------------------
-- 7. Investment Styles
-- ---------------------------------------------------------------------------
CREATE TABLE investment__investment_styles (
    code            VARCHAR(40)  PRIMARY KEY,
    name            VARCHAR(120) NOT NULL,
    description     TEXT,
    sort_order      SMALLINT     NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_investment_investment_styles_code
        CHECK (code = UPPER(code) AND code !~ '\s')
);

CREATE INDEX idx_investment_investment_styles_active ON investment__investment_styles (is_active, sort_order);

CREATE TRIGGER trg_investment_investment_styles_updated_at
    BEFORE UPDATE ON investment__investment_styles
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__investment_styles IS 'Investment style classification (GROWTH, VALUE, BLEND, INCOME, INDEX).';
