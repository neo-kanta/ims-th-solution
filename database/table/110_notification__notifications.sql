-- Table: notification__notifications
-- Source: 20260601000002_notification__create_notifications.up.sql
CREATE TABLE IF NOT EXISTS notification__notifications (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_user_id UUID         NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    category          VARCHAR(40)  NOT NULL,
    title             VARCHAR(255) NOT NULL,
    body              TEXT,
    link              TEXT,

    source_module     VARCHAR(40),
    source_type       VARCHAR(80),
    source_id         UUID,

    event_type         VARCHAR(80)  NOT NULL,
    idempotency_key    VARCHAR(180),
    business_type      VARCHAR(80),
    business_label     VARCHAR(160) NOT NULL DEFAULT '',
    business_id        UUID,
    business_reference VARCHAR(120) NOT NULL DEFAULT '',
    business_title     VARCHAR(255) NOT NULL DEFAULT '',
    action_label       VARCHAR(120) NOT NULL DEFAULT '',
    action_url         TEXT         NOT NULL DEFAULT '',

    is_read           BOOLEAN      NOT NULL DEFAULT FALSE,
    read_at           TIMESTAMPTZ,

    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
