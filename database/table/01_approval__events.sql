-- Table: approval__events
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__events (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_request_id    UUID         NOT NULL REFERENCES approval__requests(id) ON DELETE CASCADE,
    event_type             VARCHAR(30)  NOT NULL,
    stage_number           INTEGER,
    actor_user_id          UUID         REFERENCES iam_users(id),
    delegated_from_user_id UUID         REFERENCES iam_users(id),
    comment                TEXT         NOT NULL DEFAULT '',
    metadata_json          JSONB        NOT NULL DEFAULT '{}',
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_approval_event_type CHECK (event_type IN (
        'SUBMITTED','TASK_CREATED','APPROVED','REJECTED','CANCELLED','WITHDRAWN',
        'DELEGATED','STAGE_COMPLETED','REQUEST_COMPLETED'
    ))
);
