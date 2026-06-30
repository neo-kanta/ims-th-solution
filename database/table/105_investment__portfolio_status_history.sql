-- Table: investment__portfolio_status_history
-- Source: 20260613000003_investment__create_portfolio_status_history.up.sql
CREATE TABLE investment__portfolio_status_history (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id  UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,

    from_status   VARCHAR(20),
    to_status     VARCHAR(20)  NOT NULL,

    actor_id      UUID         REFERENCES iam_users(id),
    reason        TEXT,

    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_inv_psh_to_status CHECK (to_status IN (
        'DRAFT', 'PENDING_APPROVAL', 'ACTIVE', 'SUSPENDED', 'CLOSED', 'REJECTED'
    )),
    CONSTRAINT chk_inv_psh_from_status CHECK (from_status IS NULL OR from_status IN (
        'DRAFT', 'PENDING_APPROVAL', 'ACTIVE', 'PAUSED', 'SUSPENDED', 'CLOSED', 'REJECTED'
    ))
);
