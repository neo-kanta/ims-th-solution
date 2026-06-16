-- Table: investment__aum_snapshots
-- Source: 20260428000005_investment__create_pricing_valuation.up.sql
CREATE TABLE investment__aum_snapshots (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    scope_type      VARCHAR(10)  NOT NULL,
    scope_id        UUID         NOT NULL,

    business_date   DATE         NOT NULL,
    aum             DECIMAL(28,8) NOT NULL,
    valuation_ccy   CHAR(3)      NOT NULL,
    source          VARCHAR(20)  NOT NULL DEFAULT 'INTERNAL',

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by      UUID         NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_aum_scope_type
        CHECK (scope_type IN ('PORTFOLIO','FUND')),

    CONSTRAINT chk_inv_aum_ccy
        CHECK (valuation_ccy ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_aum_source
        CHECK (source IN ('INTERNAL','PAM','EXTERNAL'))
);
