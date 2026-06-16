-- Table: approval__group_members
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__group_members (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id       UUID         NOT NULL REFERENCES approval__groups(id) ON DELETE CASCADE,
    user_id        UUID         NOT NULL REFERENCES iam_users(id) ON DELETE RESTRICT,
    priority_order INTEGER      NOT NULL DEFAULT 1,
    member_type    VARCHAR(20)  NOT NULL DEFAULT 'MEMBER',
    status         VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    is_active      BOOLEAN      NOT NULL DEFAULT true,
    created_by     UUID         REFERENCES iam_users(id),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by     UUID         REFERENCES iam_users(id),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_group_members UNIQUE (group_id, user_id),
    CONSTRAINT chk_approval_group_member_type   CHECK (member_type IN ('MEMBER','SUPERVISOR')),
    CONSTRAINT chk_approval_group_member_status CHECK (status IN ('PENDING','APPROVED','REVOKED')),
    CONSTRAINT chk_approval_group_member_priority CHECK (priority_order >= 0)
);
