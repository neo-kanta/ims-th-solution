-- Table: investment__funds
-- Source: 20260428000002_investment__create_funds_portfolios.up.sql
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
    require_pretrade_preview BOOLEAN NOT NULL DEFAULT false,
    require_research_report_for_decision BOOLEAN NOT NULL DEFAULT false,
    external_pam_ref    VARCHAR(80),
    contract_code       VARCHAR(40),

    status              VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE',
    version             INTEGER      NOT NULL DEFAULT 1,

    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by          UUID         REFERENCES iam_users(id),
    updated_by          UUID         REFERENCES iam_users(id),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT chk_inv_funds_status
        CHECK (status IN (
            'DRAFT',
            'PENDING_APPROVAL',
            'ACTIVE',
            'SUSPENDED',
            'CLOSED',
            'REJECTED'
        )),

    CONSTRAINT chk_inv_funds_risk_profile
        CHECK (risk_profile IS NULL OR risk_profile IN ('LOW','MEDIUM','HIGH','SPECULATIVE')),

    CONSTRAINT chk_inv_funds_base_currency
        CHECK (base_currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_funds_version_positive
        CHECK (version >= 1)
);
