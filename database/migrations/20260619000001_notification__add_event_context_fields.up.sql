-- Add event type and business context columns to the in-app notifications table.
-- Existing rows are backfilled: event_type = category.

ALTER TABLE notification__notifications
    ADD COLUMN IF NOT EXISTS event_type         VARCHAR(80),
    ADD COLUMN IF NOT EXISTS idempotency_key    VARCHAR(180),
    ADD COLUMN IF NOT EXISTS business_type      VARCHAR(80),
    ADD COLUMN IF NOT EXISTS business_label     VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS business_id        UUID,
    ADD COLUMN IF NOT EXISTS business_reference VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS business_title     VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS action_label       VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS action_url         TEXT         NOT NULL DEFAULT '';

-- Back-fill event_type from the existing category column so existing rows
-- remain valid after the NOT NULL constraint is added below.
UPDATE notification__notifications
   SET event_type = category
 WHERE event_type IS NULL OR event_type = '';

ALTER TABLE notification__notifications
    ALTER COLUMN event_type SET NOT NULL;

-- Partial unique index: only enforce uniqueness when an idempotency key is present.
CREATE UNIQUE INDEX IF NOT EXISTS uq_notification_notifications_idempotency
    ON notification__notifications (idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';
