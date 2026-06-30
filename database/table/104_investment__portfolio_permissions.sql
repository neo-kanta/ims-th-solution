-- Table: investment__portfolio_permissions
-- Source: 20260613000005_investment__create_portfolio_permissions.up.sql
CREATE TABLE investment__portfolio_permissions (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id  UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    user_id       UUID         NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,

    role_code     VARCHAR(40)  NOT NULL,

    granted_by    UUID         NOT NULL REFERENCES iam_users(id),
    granted_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    revoked_at    TIMESTAMPTZ,
    revoked_by    UUID         REFERENCES iam_users(id),

    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_inv_pp_role_code CHECK (role_code IN (
        'PORTFOLIO_MANAGER',
        'PORTFOLIO_ANALYST',
        'PORTFOLIO_TRADER',
        'PORTFOLIO_VIEWER'
    ))
);
