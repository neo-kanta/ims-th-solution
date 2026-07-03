-- Table: approval__team_members
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__team_members (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id        UUID         NOT NULL REFERENCES approval__teams(id) ON DELETE CASCADE,
    user_id        UUID         NOT NULL REFERENCES iam_users(id) ON DELETE RESTRICT,
    member_type    VARCHAR(20)  NOT NULL DEFAULT 'REVIEWER_AGENT',
    priority_order INTEGER      NOT NULL DEFAULT 1,
    is_active      BOOLEAN      NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_team_members UNIQUE (team_id, user_id),
    CONSTRAINT chk_approval_team_member_type
        CHECK (member_type IN ('ORDER_SUBMITTER','REVIEWER_AGENT')),
    CONSTRAINT chk_approval_team_member_priority CHECK (priority_order >= 0)
);
