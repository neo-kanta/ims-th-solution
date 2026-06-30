-- Table: approval__delegations
-- Source: 20260616000002_approval__delegations.up.sql
CREATE TABLE IF NOT EXISTS approval__delegations (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id    UUID        NOT NULL,
    to_user_id      UUID        NOT NULL,
    contract_id     UUID,
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
