-- =============================================================================
-- Investment Module — Reference / Classification Tables
-- =============================================================================
-- Investment-domain taxonomy that drives instrument and fund classification.
-- DB-driven on purpose: adding a new asset class / sector / fund category must
-- not require a code change. Application-layer dispatch (e.g. per asset-class
-- valuator) keys on the stable `code` column, not the row UUID.
--
-- Tables created here:
--   investment__asset_classes
--   investment__asset_subtypes
--   investment__regions
--   investment__countries
--   investment__sectors          (self-referencing for GICS-style hierarchy)
--   investment__fund_categories
--   investment__investment_styles
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. Asset classes
-- ---------------------------------------------------------------------------
CREATE TABLE investment__asset_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    code VARCHAR(40) NOT NULL,
    name VARCHAR(120) NOT NULL,
    description TEXT,
    display_order SMALLINT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_asset_classes_code UNIQUE (code)
);

CREATE INDEX idx_inv_asset_classes_active ON investment__asset_classes (is_active, display_order);

CREATE TRIGGER trg_inv_asset_classes_updated_at
    BEFORE UPDATE ON investment__asset_classes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON
TABLE investment__asset_classes IS 'Top-level asset class taxonomy (EQUITY, FIXED_INCOME, FUND, ETF, CASH, ALTERNATIVE, DERIVATIVE).';

COMMENT ON COLUMN investment__asset_classes.code IS 'Stable application-side dispatch key. Must not be renamed without coordinated code change.';

-- ---------------------------------------------------------------------------
-- 2. Asset subtypes
-- ---------------------------------------------------------------------------
CREATE TABLE investment__asset_subtypes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    asset_class_id UUID NOT NULL REFERENCES investment__asset_classes (id) ON DELETE RESTRICT,
    code VARCHAR(40) NOT NULL,
    name VARCHAR(120) NOT NULL,
    description TEXT,
    display_order SMALLINT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_asset_subtypes_class_code UNIQUE (asset_class_id, code)
);

CREATE INDEX idx_inv_asset_subtypes_class ON investment__asset_subtypes (asset_class_id, is_active);

CREATE TRIGGER trg_inv_asset_subtypes_updated_at
    BEFORE UPDATE ON investment__asset_subtypes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON
TABLE investment__asset_subtypes IS 'Sub-classification under an asset class (COMMON_STOCK, EQUITY_ETF, MUTUAL_FUND, GOV_BOND, etc.).';

-- ---------------------------------------------------------------------------
-- 3. Regions
-- ---------------------------------------------------------------------------
CREATE TABLE investment__regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    code VARCHAR(20) NOT NULL,
    name VARCHAR(120) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_regions_code UNIQUE (code)
);

CREATE TRIGGER trg_inv_regions_updated_at
    BEFORE UPDATE ON investment__regions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON
TABLE investment__regions IS 'Macro investment regions (APAC, EMEA, AMER, GLOBAL, THAILAND, etc.).';

-- ---------------------------------------------------------------------------
-- 4. Countries (ISO 3166-1 alpha-2)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__countries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    iso_code CHAR(2) NOT NULL,
    name VARCHAR(120) NOT NULL,
    region_id UUID REFERENCES investment__regions (id) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_countries_iso UNIQUE (iso_code),
    CONSTRAINT chk_inv_countries_iso_format CHECK (iso_code ~ '^[A-Z]{2}$')
);

CREATE INDEX idx_inv_countries_region ON investment__countries (region_id, is_active);

CREATE TRIGGER trg_inv_countries_updated_at
    BEFORE UPDATE ON investment__countries
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON
TABLE investment__countries IS 'Investment-context country list. Independent of any system-level locale data.';

-- ---------------------------------------------------------------------------
-- 5. Sectors (GICS-style hierarchy via self FK)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__sectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    code VARCHAR(40) NOT NULL,
    name VARCHAR(120) NOT NULL,
    parent_id UUID REFERENCES investment__sectors (id) ON DELETE RESTRICT,
    level SMALLINT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_sectors_parent_code UNIQUE (parent_id, code),
    CONSTRAINT chk_inv_sectors_level CHECK (level BETWEEN 1 AND 4)
);

CREATE INDEX idx_inv_sectors_parent ON investment__sectors (parent_id, level);

CREATE INDEX idx_inv_sectors_level ON investment__sectors (level, is_active);

CREATE TRIGGER trg_inv_sectors_updated_at
    BEFORE UPDATE ON investment__sectors
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON
TABLE investment__sectors IS 'GICS-compatible sector hierarchy (Sector -> Industry Group -> Industry -> Sub-Industry).';

COMMENT ON COLUMN investment__sectors.level IS '1=Sector, 2=Industry Group, 3=Industry, 4=Sub-Industry.';

-- ---------------------------------------------------------------------------
-- 6. Fund categories
-- ---------------------------------------------------------------------------
CREATE TABLE investment__fund_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    code VARCHAR(40) NOT NULL,
    name VARCHAR(120) NOT NULL,
    asset_class_id UUID REFERENCES investment__asset_classes (id) ON DELETE RESTRICT,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_fund_categories_code UNIQUE (code)
);

CREATE INDEX idx_inv_fund_categories_class ON investment__fund_categories (asset_class_id, is_active);

CREATE TRIGGER trg_inv_fund_categories_updated_at
    BEFORE UPDATE ON investment__fund_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON
TABLE investment__fund_categories IS 'Fund / portfolio category (EQUITY_FUND, BOND_FUND, MIXED_FUND, MMF_FUND, INDEX_FUND, ...).';

-- ---------------------------------------------------------------------------
-- 7. Investment styles (optional metadata)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__investment_styles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    code VARCHAR(40) NOT NULL,
    name VARCHAR(120) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_styles_code UNIQUE (code)
);

CREATE TRIGGER trg_inv_investment_styles_updated_at
    BEFORE UPDATE ON investment__investment_styles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON
TABLE investment__investment_styles IS 'Investment style classification (GROWTH, VALUE, BLEND, INCOME, INDEX, ACTIVE).';