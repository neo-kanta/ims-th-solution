-- Table: approval_workflow_settings
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE approval_workflow_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    setting_code VARCHAR(100) NOT NULL,
    setting_name VARCHAR(180) NOT NULL,
    module VARCHAR(80) NOT NULL,
    request_type VARCHAR(60) NOT NULL,
    risk_level VARCHAR(20) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    sequential_approval BOOLEAN NOT NULL DEFAULT true,
    allow_creator_approval BOOLEAN NOT NULL DEFAULT false,
    failed_check_blocks_submit BOOLEAN NOT NULL DEFAULT true,
    failed_check_blocks_merge BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_workflow_settings_code UNIQUE (setting_code),
    CONSTRAINT chk_approval_workflow_settings_risk CHECK (risk_level IN (
        'LOW','MEDIUM','HIGH','CRITICAL'
    ))
);
