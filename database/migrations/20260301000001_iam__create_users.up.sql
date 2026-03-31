-- IAM Users table (clean IAM boundary — no employee/HR state)
-- Uses gen_random_uuid() (pg14+ built-in, no extension needed)

CREATE TABLE IF NOT EXISTS iam_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    force_password_change BOOLEAN NOT NULL DEFAULT false,
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_iam_users_username UNIQUE (username)
);

CREATE INDEX idx_iam_users_username ON iam_users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_iam_users_is_active ON iam_users(is_active) WHERE deleted_at IS NULL;

COMMENT ON TABLE iam_users IS 'User accounts for authentication and access management';
COMMENT ON COLUMN iam_users.force_password_change IS 'When true, user must change password on next login';
COMMENT ON COLUMN iam_users.failed_login_attempts IS 'Consecutive failed login attempts; reset on success';
COMMENT ON COLUMN iam_users.locked_until IS 'Account locked until this timestamp; NULL means not locked';
COMMENT ON COLUMN iam_users.version IS 'Optimistic locking version counter';
COMMENT ON COLUMN iam_users.deleted_at IS 'Soft delete timestamp; NULL means active record';

-- Trigger: auto-update updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_iam_users_updated_at
    BEFORE UPDATE ON iam_users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
