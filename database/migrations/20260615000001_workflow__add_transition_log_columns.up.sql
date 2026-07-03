ALTER TABLE workflow__transition_log
    ADD COLUMN actor_account_code VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN is_admin_override   BOOLEAN      NOT NULL DEFAULT FALSE;
