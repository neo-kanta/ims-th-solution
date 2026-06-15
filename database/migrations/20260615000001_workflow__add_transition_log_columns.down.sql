ALTER TABLE workflow__transition_log
    DROP COLUMN IF EXISTS actor_account_code,
    DROP COLUMN IF EXISTS is_admin_override;
