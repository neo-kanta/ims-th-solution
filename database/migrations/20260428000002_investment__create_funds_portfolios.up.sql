-- =============================================================================
-- Investment Module — Fund (Contract) Master + Portfolio Master
-- =============================================================================
-- A Fund is the legal vehicle. Its `id` IS the cross-module `contract_id`
-- consumed by workflow / compliance / permissions modules. We do not maintain
-- a separate contract table — the fund row is the contract.
--
-- A Portfolio belongs to exactly one fund (1..*). Portfolios may differ from
-- the parent fund in valuation currency or strategy (sleeves).
--
-- Soft delete: master tables use `deleted_at`. Application-level reads MUST
-- filter `deleted_at IS NULL`.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. Funds (== contracts)
-- ---------------------------------------------------------------------------
CREATE TABLE investment__funds (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    code                VARCHAR(40)  NOT NULL,
    name                VARCHAR(255) NOT NULL,
    short_name          VARCHAR(80),

    fund_category_id    UUID         NOT NULL REFERENCES investment__fund_categories(id) ON DELETE RESTRICT,
    base_currency       CHAR(3)      NOT NULL,

    inception_date      DATE         NOT NULL,
    manager_user_id     UUID         REFERENCES iam_users(id),

    benchmark           VARCHAR(120),
    risk_profile        VARCHAR(20),
    has_units           BOOLEAN      NOT NULL DEFAULT false,
    external_pam_ref    VARCHAR(80),

    status              VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE',
    version             INTEGER      NOT NULL DEFAULT 1,

    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by          UUID         REFERENCES iam_users(id),
    updated_by          UUID         REFERENCES iam_users(id),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT chk_inv_funds_status
        CHECK (status IN ('ACTIVE', 'SUSPENDED', 'CLOSED')),

    CONSTRAINT chk_inv_funds_risk_profile
        CHECK (risk_profile IS NULL OR risk_profile IN ('LOW','MEDIUM','HIGH','SPECULATIVE')),

    CONSTRAINT chk_inv_funds_base_currency
        CHECK (base_currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_funds_version_positive
        CHECK (version >= 1)
);

CREATE UNIQUE INDEX uq_inv_funds_code_alive
    ON investment__funds (code) WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_funds_status      ON investment__funds (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_inv_funds_category    ON investment__funds (fund_category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_inv_funds_manager     ON investment__funds (manager_user_id) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_inv_funds_updated_at
    BEFORE UPDATE ON investment__funds
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE  investment__funds IS 'Fund / contract master. Fund.id is the cross-module contract_id.';
COMMENT ON COLUMN investment__funds.has_units IS 'True for unitised funds where NAV-per-unit is meaningful.';
COMMENT ON COLUMN investment__funds.external_pam_ref IS 'Future PAM linkage; nullable for PoC.';

-- ---------------------------------------------------------------------------
-- 2. Portfolios
-- ---------------------------------------------------------------------------
CREATE TABLE investment__portfolios (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    fund_id             UUID         NOT NULL REFERENCES investment__funds(id) ON DELETE RESTRICT,
    code                VARCHAR(40)  NOT NULL,
    name                VARCHAR(255) NOT NULL,
    description         TEXT,

    base_currency       CHAR(3)      NOT NULL,
    valuation_currency  CHAR(3)      NOT NULL,

    strategy_code       VARCHAR(40),
    style_id            UUID         REFERENCES investment__investment_styles(id) ON DELETE RESTRICT,

    manager_user_id     UUID         REFERENCES iam_users(id),
    benchmark           VARCHAR(120),
    risk_profile        VARCHAR(20),
    inception_date      DATE         NOT NULL,

    status              VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE',
    has_units           BOOLEAN      NOT NULL DEFAULT false,
    tax_lot_method      VARCHAR(20)  NOT NULL DEFAULT 'AVERAGE',

    version             INTEGER      NOT NULL DEFAULT 1,

    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by          UUID         REFERENCES iam_users(id),
    updated_by          UUID         REFERENCES iam_users(id),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT chk_inv_portfolios_status
        CHECK (status IN ('ACTIVE', 'PAUSED', 'CLOSED')),

    CONSTRAINT chk_inv_portfolios_risk_profile
        CHECK (risk_profile IS NULL OR risk_profile IN ('LOW','MEDIUM','HIGH','SPECULATIVE')),

    CONSTRAINT chk_inv_portfolios_tax_lot_method
        CHECK (tax_lot_method IN ('AVERAGE','FIFO','LIFO','SPEC_ID')),

    CONSTRAINT chk_inv_portfolios_base_currency
        CHECK (base_currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_portfolios_valuation_currency
        CHECK (valuation_currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_portfolios_version_positive
        CHECK (version >= 1)
);

CREATE UNIQUE INDEX uq_inv_portfolios_fund_code_alive
    ON investment__portfolios (fund_id, code) WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_portfolios_fund     ON investment__portfolios (fund_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_inv_portfolios_manager  ON investment__portfolios (manager_user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_inv_portfolios_status   ON investment__portfolios (status) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_inv_portfolios_updated_at
    BEFORE UPDATE ON investment__portfolios
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE  investment__portfolios IS 'Portfolio master. Belongs to exactly one fund.';
COMMENT ON COLUMN investment__portfolios.tax_lot_method IS 'Placeholder for future cost methods. Implementation today is AVERAGE only.';
COMMENT ON COLUMN investment__portfolios.has_units IS 'When true, NAV-per-unit and unit issuance/redemption ledger entries apply.';
