-- Table: investment__portfolios
-- Source: 20260428000002_investment__create_funds_portfolios.up.sql
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
        CHECK (status IN (
            'DRAFT',
            'PENDING_APPROVAL',
            'ACTIVE',
            'PAUSED',
            'SUSPENDED',
            'CLOSED',
            'REJECTED'
        )),

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
