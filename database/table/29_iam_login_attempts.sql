-- Table: iam_login_attempts
-- Source: 20260301000006_iam__add_mfa_and_session_enhancements.up.sql
CREATE TABLE IF NOT EXISTS iam_login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address INET NOT NULL,
    username VARCHAR(100),
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL DEFAULT false
);
