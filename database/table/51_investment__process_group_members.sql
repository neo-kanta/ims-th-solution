-- Table: investment__process_group_members
-- Source: 20260427000002_investment__create_process_assignment_tables.up.sql
CREATE TABLE investment__process_group_members (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    group_id        UUID        NOT NULL REFERENCES investment__process_groups(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,

    member_role     VARCHAR(30) NOT NULL DEFAULT 'MEMBER',
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    effective_from  DATE        NOT NULL,
    effective_to    DATE,

    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by     UUID        REFERENCES iam_users(id),

    CONSTRAINT uq_investment_process_group_member UNIQUE (group_id, user_id),

    CONSTRAINT chk_investment_process_group_member_role
        CHECK (member_role IN ('LEAD', 'MEMBER', 'BACKUP')),

    CONSTRAINT chk_investment_process_group_member_effective_range
        CHECK (effective_to IS NULL OR effective_to >= effective_from)
);
