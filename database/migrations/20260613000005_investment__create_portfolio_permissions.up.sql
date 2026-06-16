-- =============================================================================
-- Investment Module — Portfolio data-access permissions table
-- =============================================================================
-- Portfolio permissions are a second-layer data-access control on top of
-- IAM function-level RBAC. They answer: "which portfolios can this user see
-- and act on?" — not "what actions are they allowed to perform?" (that is
-- the IAM RBAC question).
--
-- A user can hold at most one ACTIVE (non-revoked) grant per portfolio.
-- The partial unique index below enforces this at the database level.
-- Re-granting after revocation is allowed — it creates a new row.
--
-- Role codes:
--   PORTFOLIO_MANAGER   — full read/write, can initiate lifecycle transitions
--   PORTFOLIO_ANALYST   — read + can create research reports and decisions
--   PORTFOLIO_TRADER    — read + can post ledger transactions
--   PORTFOLIO_VIEWER    — read-only
-- =============================================================================

CREATE TABLE investment__portfolio_permissions (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id  UUID         NOT NULL
                      REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    user_id       UUID         NOT NULL
                      REFERENCES iam_users(id)              ON DELETE CASCADE,

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

-- At most one active (non-revoked) grant per (portfolio, user).
CREATE UNIQUE INDEX uq_inv_pp_active_grant
    ON investment__portfolio_permissions (portfolio_id, user_id)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_inv_pp_portfolio
    ON investment__portfolio_permissions (portfolio_id)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_inv_pp_user
    ON investment__portfolio_permissions (user_id)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_inv_pp_granted_by
    ON investment__portfolio_permissions (granted_by);

CREATE TRIGGER trg_inv_pp_updated_at
    BEFORE UPDATE ON investment__portfolio_permissions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE  investment__portfolio_permissions IS 'Data-access grants that control which users can see and act on a specific portfolio.';
COMMENT ON COLUMN investment__portfolio_permissions.role_code IS 'PORTFOLIO_MANAGER | PORTFOLIO_ANALYST | PORTFOLIO_TRADER | PORTFOLIO_VIEWER';
COMMENT ON COLUMN investment__portfolio_permissions.revoked_at IS 'NULL = active grant. Set by revoke operation; row is kept for audit.';
