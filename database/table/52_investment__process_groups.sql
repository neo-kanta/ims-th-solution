-- Table: investment__process_groups
-- Source: 20260427000002_investment__create_process_assignment_tables.up.sql
CREATE TABLE investment__process_groups (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    group_key       VARCHAR(80) NOT NULL,
    name            VARCHAR(120) NOT NULL,
    description     TEXT,

    is_active       BOOLEAN     NOT NULL DEFAULT true,
    effective_from  DATE        NOT NULL,
    effective_to    DATE,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      UUID        REFERENCES iam_users(id),
    updated_by      UUID        REFERENCES iam_users(id),

    CONSTRAINT uq_investment_process_groups_key UNIQUE (group_key),
    CONSTRAINT uq_investment_process_groups_name UNIQUE (name),

    CONSTRAINT chk_investment_process_groups_effective_range
        CHECK (effective_to IS NULL OR effective_to >= effective_from)
);
