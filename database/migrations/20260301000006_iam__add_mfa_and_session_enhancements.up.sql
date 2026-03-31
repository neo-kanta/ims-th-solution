-- IAM Enhancement: MFA (TOTP), session idle tracking, password expiration, rate limiting
-- Migration: 20260301000006

-- =============================================================================
-- 1. MFA (TOTP) — Separate table, not polluting iam_users
-- =============================================================================

CREATE TABLE IF NOT EXISTS iam_mfa_enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    mfa_type VARCHAR(20) NOT NULL DEFAULT 'totp',
    secret_encrypted VARCHAR(512) NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    is_enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ,
    CONSTRAINT uq_iam_mfa_user_type UNIQUE (user_id, mfa_type)
);

CREATE INDEX idx_iam_mfa_user ON iam_mfa_enrollments(user_id) WHERE is_enabled = true;

COMMENT ON TABLE iam_mfa_enrollments IS 'MFA enrollment state per user; secret_encrypted stores the TOTP shared secret (encrypted at rest)';
COMMENT ON COLUMN iam_mfa_enrollments.secret_encrypted IS 'TOTP shared secret encrypted with app-level key; never store plaintext';
COMMENT ON COLUMN iam_mfa_enrollments.is_verified IS 'True after user confirms first TOTP code during enrollment';
COMMENT ON COLUMN iam_mfa_enrollments.is_enabled IS 'True when MFA is active for this user (verified + enabled)';

-- =============================================================================
-- 2. MFA Recovery Codes — hashed, single-use
-- =============================================================================

CREATE TABLE IF NOT EXISTS iam_mfa_recovery_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    is_used BOOLEAN NOT NULL DEFAULT false,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_iam_mfa_recovery_user ON iam_mfa_recovery_codes(user_id) WHERE is_used = false;

COMMENT ON TABLE iam_mfa_recovery_codes IS 'Hashed single-use recovery codes for MFA bypass';

-- =============================================================================
-- 3. Session enhancements — idle tracking + absolute lifetime
-- =============================================================================

ALTER TABLE iam_sessions
    ADD COLUMN IF NOT EXISTS last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS absolute_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(255),
    ADD COLUMN IF NOT EXISTS revoke_reason VARCHAR(100);

-- Backfill absolute_expires_at for existing sessions
UPDATE iam_sessions
SET absolute_expires_at = expires_at,
    last_activity_at = COALESCE(rotated_at, created_at)
WHERE absolute_expires_at IS NULL;

-- Make absolute_expires_at NOT NULL after backfill
ALTER TABLE iam_sessions ALTER COLUMN absolute_expires_at SET NOT NULL;
ALTER TABLE iam_sessions ALTER COLUMN absolute_expires_at SET DEFAULT NOW() + INTERVAL '7 days';

CREATE INDEX idx_iam_sessions_last_activity ON iam_sessions(last_activity_at) WHERE is_revoked = false;

COMMENT ON COLUMN iam_sessions.last_activity_at IS 'Last time this session was used (for idle timeout enforcement)';
COMMENT ON COLUMN iam_sessions.absolute_expires_at IS 'Hard session lifetime limit regardless of activity';
COMMENT ON COLUMN iam_sessions.revoke_reason IS 'Why session was revoked: idle_timeout, admin_revoke, logout, breach, password_change, concurrent_limit';

-- =============================================================================
-- 4. Password expiration tracking on iam_users
-- =============================================================================

ALTER TABLE iam_users
    ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ;

-- Backfill: assume current password was set at account creation
UPDATE iam_users SET password_changed_at = created_at WHERE password_changed_at IS NULL;

ALTER TABLE iam_users ALTER COLUMN password_changed_at SET DEFAULT NOW();

COMMENT ON COLUMN iam_users.password_changed_at IS 'When the password was last changed; used for expiration policy enforcement';

-- =============================================================================
-- 5. Login rate limiting — lightweight DB-backed (optional, app can use in-memory)
-- =============================================================================

CREATE TABLE IF NOT EXISTS iam_login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address INET NOT NULL,
    username VARCHAR(100),
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_iam_login_attempts_ip ON iam_login_attempts(ip_address, attempted_at);
CREATE INDEX idx_iam_login_attempts_cleanup ON iam_login_attempts(attempted_at);

COMMENT ON TABLE iam_login_attempts IS 'Login attempt log for rate limiting; entries older than window are periodically purged';

-- =============================================================================
-- 6. JWT signing key metadata (for key rotation)
-- =============================================================================

CREATE TABLE IF NOT EXISTS iam_signing_keys (
    kid VARCHAR(64) PRIMARY KEY,
    algorithm VARCHAR(20) NOT NULL DEFAULT 'HS256',
    key_material_encrypted VARCHAR(1024) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    activated_at TIMESTAMPTZ,
    retired_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_iam_signing_keys_active ON iam_signing_keys(is_active) WHERE is_active = true;

COMMENT ON TABLE iam_signing_keys IS 'JWT signing key metadata for key rotation; only one key is_active for signing at a time';
COMMENT ON COLUMN iam_signing_keys.kid IS 'Key ID embedded in JWT header for verification key lookup';
COMMENT ON COLUMN iam_signing_keys.key_material_encrypted IS 'Encrypted signing key; decrypted at app startup';

-- =============================================================================
-- 7. Additional audit event indexes for query API performance
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_iam_audit_events_target_id ON iam_audit_events(target_id);
CREATE INDEX IF NOT EXISTS idx_iam_audit_events_composite ON iam_audit_events(event_type, created_at DESC);
