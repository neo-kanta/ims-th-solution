DROP INDEX IF EXISTS uq_notification_notifications_idempotency;

ALTER TABLE notification__notifications
    ALTER COLUMN event_type DROP NOT NULL;

ALTER TABLE notification__notifications
    DROP COLUMN IF EXISTS event_type,
    DROP COLUMN IF EXISTS idempotency_key,
    DROP COLUMN IF EXISTS business_type,
    DROP COLUMN IF EXISTS business_label,
    DROP COLUMN IF EXISTS business_id,
    DROP COLUMN IF EXISTS business_reference,
    DROP COLUMN IF EXISTS business_title,
    DROP COLUMN IF EXISTS action_label,
    DROP COLUMN IF EXISTS action_url;
