-- Table: iam_mfa_enrollments
-- Source: 20260301000006_iam__add_mfa_and_session_enhancements.up.sql
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
