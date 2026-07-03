-- Table: iam_signing_keys
-- Source: 20260301000006_iam__add_mfa_and_session_enhancements.up.sql
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
