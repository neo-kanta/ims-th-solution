-- Table: investment__fund_status_history
-- Source: 20260613000004_investment__create_fund_status_history.up.sql
CREATE TABLE investment__fund_status_history (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    fund_id     UUID         NOT NULL REFERENCES investment__funds(id) ON DELETE RESTRICT,

    from_status VARCHAR(20),
    to_status   VARCHAR(20)  NOT NULL,

    actor_id    UUID         REFERENCES iam_users(id),
    reason      TEXT,

    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_inv_fsh_to_status CHECK (to_status IN (
        'DRAFT', 'PENDING_APPROVAL', 'ACTIVE', 'SUSPENDED', 'CLOSED', 'REJECTED'
    )),
    CONSTRAINT chk_inv_fsh_from_status CHECK (from_status IS NULL OR from_status IN (
        'DRAFT', 'PENDING_APPROVAL', 'ACTIVE', 'SUSPENDED', 'CLOSED', 'REJECTED'
    ))
);
