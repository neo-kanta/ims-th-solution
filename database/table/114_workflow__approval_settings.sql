-- Table: workflow__approval_settings
-- Source: 20260615000002_workflow__create_approval_settings.up.sql
CREATE TABLE workflow__approval_settings (
    id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    operation_type        VARCHAR(60)  NOT NULL,
    approval_mode         VARCHAR(20)  NOT NULL DEFAULT 'ANY_OF',
    approver_account_code VARCHAR(255) NOT NULL,
    approver_username     VARCHAR(255) NOT NULL DEFAULT '',
    approver_role         VARCHAR(100) NOT NULL DEFAULT '',
    is_active             BOOLEAN      NOT NULL DEFAULT TRUE,
    updated_by            VARCHAR(255) NOT NULL DEFAULT '',
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now()
);
