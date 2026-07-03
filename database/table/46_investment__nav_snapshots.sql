-- Table: investment__nav_snapshots
-- Source: 20260428000005_investment__create_pricing_valuation.up.sql
CREATE TABLE investment__nav_snapshots (
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id             UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    business_date            DATE         NOT NULL,

    total_units              DECIMAL(28,8) NOT NULL,
    nav_per_unit             DECIMAL(28,12) NOT NULL,

    valuation_snapshot_id    UUID         NOT NULL REFERENCES investment__valuation_snapshots(id) ON DELETE RESTRICT,

    is_indicative            BOOLEAN      NOT NULL DEFAULT true,
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by               UUID         NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_nav_units_non_negative
        CHECK (total_units >= 0)
);
