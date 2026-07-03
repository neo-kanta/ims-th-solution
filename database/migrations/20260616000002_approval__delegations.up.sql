-- Approval delegation table.
-- Records that a user (from_user_id) has delegated their approval authority to
-- another user (to_user_id) for a given time window. An optional
-- contract_id_filter restricts delegation to a specific contract/fund; NULL
-- means the delegation applies to all contracts.
CREATE TABLE IF NOT EXISTS approval__delegations (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id    UUID        NOT NULL,
    to_user_id      UUID        NOT NULL,
    contract_id     UUID        NULL,
    active_from     TIMESTAMPTZ NOT NULL,
    active_until    TIMESTAMPTZ NOT NULL,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    remarks         TEXT        NOT NULL DEFAULT '',
    created_by      UUID        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_delegation_window CHECK (active_until > active_from),
    CONSTRAINT chk_delegation_no_self CHECK (from_user_id <> to_user_id)
);

CREATE INDEX IF NOT EXISTS idx_approval__delegations_from_active
    ON approval__delegations (from_user_id, is_active, active_from, active_until);
