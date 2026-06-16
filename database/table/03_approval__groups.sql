-- Table: approval__groups
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__groups (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    group_code  VARCHAR(60)  NOT NULL,
    group_name  VARCHAR(160) NOT NULL,
    remarks     TEXT         NOT NULL DEFAULT '',
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_by  UUID         REFERENCES iam_users(id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by  UUID         REFERENCES iam_users(id),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_groups_code UNIQUE (group_code)
);
