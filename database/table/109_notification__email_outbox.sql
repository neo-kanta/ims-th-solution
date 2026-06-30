-- Table: notification__email_outbox
-- Source: 20260619000002_notification__email_outbox.up.sql
CREATE TABLE notification__email_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(180) NOT NULL,

    notification_id UUID REFERENCES notification__notifications(id) ON DELETE SET NULL,
    recipient_user_id UUID REFERENCES iam_users(id) ON DELETE SET NULL,

    recipient_username     VARCHAR(100)  NOT NULL DEFAULT '',
    recipient_display_name VARCHAR(255)  NOT NULL DEFAULT '',
    recipient_email        VARCHAR(255)  NOT NULL DEFAULT '',

    to_email VARCHAR(255) NOT NULL,
    to_name  VARCHAR(255) NOT NULL DEFAULT '',

    event_type     VARCHAR(80)  NOT NULL,
    event_label    VARCHAR(160) NOT NULL DEFAULT '',
    event_category VARCHAR(40)  NOT NULL DEFAULT '',
    event_severity VARCHAR(20)  NOT NULL DEFAULT 'INFO',

    business_type      VARCHAR(80),
    business_label     VARCHAR(160) NOT NULL DEFAULT '',
    business_id        UUID,
    business_reference VARCHAR(120) NOT NULL DEFAULT '',
    business_title     VARCHAR(255) NOT NULL DEFAULT '',

    action_label VARCHAR(120) NOT NULL DEFAULT '',
    action_url   TEXT         NOT NULL DEFAULT '',

    subject   VARCHAR(255) NOT NULL,
    body_text TEXT         NOT NULL,
    body_html TEXT,

    status          VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    attempts        INTEGER      NOT NULL DEFAULT 0,
    max_attempts    INTEGER      NOT NULL DEFAULT 5,
    next_attempt_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    locked_at TIMESTAMPTZ,
    locked_by VARCHAR(120),

    sent_at             TIMESTAMPTZ,
    last_error          TEXT,
    provider_message_id TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_notification_email_status CHECK (
        status IN ('PENDING', 'SENDING', 'SENT', 'FAILED', 'DEAD', 'CANCELLED')
    ),
    CONSTRAINT chk_notification_email_attempts CHECK (attempts >= 0),
    CONSTRAINT chk_notification_email_max_attempts CHECK (max_attempts > 0),
    CONSTRAINT uq_notification_email_outbox_idempotency UNIQUE (idempotency_key)
);
