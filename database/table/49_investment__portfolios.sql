-- Table: investment__portfolios
-- Source: 20260428000002_investment__create_funds_portfolios.up.sql,
--         updated by 20260613000002_investment__expand_portfolio_status_lifecycle.up.sql,
--         updated by 20260703000001_investment__portfolio_v2_hardening.up.sql
-- Purpose: Portfolio master. Portfolio is the operational/accounting source of truth.
-- Identity: Public APIs use portfolio code; database FKs use id. fund_id is still present
--           as the current fund wrapper link and is protected by composite constraints.
CREATE TABLE investment__portfolios (
    -- Identity
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    code                VARCHAR(40)  NOT NULL,
    portfolio_type      VARCHAR(20)  NOT NULL DEFAULT 'LIVE',
    name                VARCHAR(255) NOT NULL,
    description         TEXT,

    -- Optional fund wrapper / legacy compatibility
    fund_id             UUID         NOT NULL REFERENCES investment__funds(id) ON DELETE RESTRICT,

    -- Business attributes
    base_currency       CHAR(3)      NOT NULL,
    valuation_currency  CHAR(3)      NOT NULL,
    strategy_code       VARCHAR(40),
    style_id            UUID         REFERENCES investment__investment_styles(id) ON DELETE RESTRICT,
    manager_user_id     UUID         REFERENCES iam_users(id),
    benchmark           VARCHAR(120),
    risk_profile        VARCHAR(20),
    inception_date      DATE         NOT NULL,
    has_units           BOOLEAN      NOT NULL DEFAULT false,
    tax_lot_method      VARCHAR(20)  NOT NULL DEFAULT 'AVERAGE',

    -- Lifecycle / status
    status              VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE',
    version             INTEGER      NOT NULL DEFAULT 1,

    -- Audit
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by          UUID         REFERENCES iam_users(id),
    updated_by          UUID         REFERENCES iam_users(id),
    deleted_at          TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT uq_inv_portfolios_id_fund
        UNIQUE (id, fund_id),

    CONSTRAINT chk_inv_portfolios_status
        CHECK (status IN (
            'DRAFT',
            'PENDING_APPROVAL',
            'ACTIVE',
            'PAUSED',
            'SUSPENDED',
            'CLOSED',
            'REJECTED'
        )),

    CONSTRAINT chk_inv_portfolios_type
        CHECK (portfolio_type IN ('LIVE', 'SIMULATION', 'MODEL')),

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

-- Indexes
CREATE UNIQUE INDEX uq_inv_portfolios_fund_code_alive
    ON investment__portfolios (fund_id, code) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_inv_portfolios_code_alive
    ON investment__portfolios (code) WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_portfolios_fund
    ON investment__portfolios (fund_id, status) WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_portfolios_manager
    ON investment__portfolios (manager_user_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_portfolios_status
    ON investment__portfolios (status) WHERE deleted_at IS NULL;

-- Triggers
CREATE TRIGGER trg_inv_portfolios_updated_at
    BEFORE UPDATE ON investment__portfolios
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Comments
COMMENT ON TABLE  investment__portfolios IS
    'Portfolio master. Operational source of truth; fund_id is the current required wrapper link.';
COMMENT ON COLUMN investment__portfolios.portfolio_type IS
    'Portfolio V2 type (docs/api/portfolio-v2-api-ddd.md section 4). Existing rows default to LIVE.';
COMMENT ON COLUMN investment__portfolios.tax_lot_method IS
    'Placeholder for future cost methods. Implementation today is AVERAGE only.';
COMMENT ON COLUMN investment__portfolios.has_units IS
    'When true, NAV-per-unit and unit issuance/redemption ledger entries apply.';
