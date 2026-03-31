-- Rollback: IAM Enhancement (MFA, session idle, password expiry, rate limiting)

DROP INDEX IF EXISTS idx_iam_audit_events_composite;
DROP INDEX IF EXISTS idx_iam_audit_events_target_id;

DROP TABLE IF EXISTS iam_signing_keys;
DROP TABLE IF EXISTS iam_login_attempts;
DROP TABLE IF EXISTS iam_mfa_recovery_codes;
DROP TABLE IF EXISTS iam_mfa_enrollments;

DROP INDEX IF EXISTS idx_iam_sessions_last_activity;

ALTER TABLE iam_sessions DROP COLUMN IF EXISTS revoke_reason;
ALTER TABLE iam_sessions DROP COLUMN IF EXISTS device_fingerprint;
ALTER TABLE iam_sessions DROP COLUMN IF EXISTS absolute_expires_at;
ALTER TABLE iam_sessions DROP COLUMN IF EXISTS last_activity_at;

ALTER TABLE iam_users DROP COLUMN IF EXISTS password_changed_at;
