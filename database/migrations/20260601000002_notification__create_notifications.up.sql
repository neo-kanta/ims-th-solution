-- =============================================================================
-- Notification baseline: in-app notifications for approval events.
-- =============================================================================
--
-- Scope is intentionally narrow for the demo: store a row per recipient when
-- an approval task is assigned, completed, or rejected. The frontend lists
-- the current user's unread/read notifications and marks them read. Email /
-- push channels are out of scope for this phase.
-- =============================================================================

CREATE TABLE IF NOT EXISTS notification__notifications (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_user_id UUID          NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    category        VARCHAR(40)     NOT NULL,
    title           VARCHAR(255)    NOT NULL,
    body            TEXT,
    link            TEXT,

    -- Polymorphic reference to whatever business object generated the
    -- notification. Kept as plain strings (no FK) so the table never blocks
    -- writes from any source module.
    source_module   VARCHAR(40),
    source_type     VARCHAR(80),
    source_id       UUID,

    is_read         BOOLEAN         NOT NULL DEFAULT FALSE,
    read_at         TIMESTAMPTZ,

    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_recipient_unread
    ON notification__notifications (recipient_user_id, is_read, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notification_recipient_created
    ON notification__notifications (recipient_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notification_source
    ON notification__notifications (source_module, source_type, source_id)
    WHERE source_id IS NOT NULL;

COMMENT ON TABLE notification__notifications IS 'Per-recipient in-app notification log. Minimum baseline; email / push out of scope.';
