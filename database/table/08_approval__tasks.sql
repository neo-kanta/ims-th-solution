-- Table: approval__tasks
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__tasks (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_request_id    UUID         NOT NULL REFERENCES approval__requests(id) ON DELETE CASCADE,
    stage_number           INTEGER      NOT NULL,
    assigned_user_id       UUID         REFERENCES iam_users(id),
    assigned_group_id      UUID         REFERENCES approval__groups(id),
    assigned_team_id       UUID         REFERENCES approval__teams(id),
    delegated_from_user_id UUID         REFERENCES iam_users(id),
    status                 VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    acted_by               UUID         REFERENCES iam_users(id),
    acted_at               TIMESTAMPTZ,
    action_comment         TEXT         NOT NULL DEFAULT '',
    is_delegated_action    BOOLEAN      NOT NULL DEFAULT false,
    due_at                 TIMESTAMPTZ,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_approval_task_status CHECK (status IN (
        'PENDING','APPROVED','REJECTED','SKIPPED','CANCELLED'
    )),
    CONSTRAINT chk_approval_task_stage CHECK (stage_number BETWEEN 1 AND 3)
);
