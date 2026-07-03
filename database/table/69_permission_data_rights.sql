-- Table: permission_data_rights
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_data_rights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_type VARCHAR(20) NOT NULL,
    subject_id UUID NOT NULL,
    fund_id UUID REFERENCES investment__funds(id) ON DELETE RESTRICT,
    contract_id UUID REFERENCES investment__funds(id) ON DELETE RESTRICT,
    portfolio_id UUID REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    access_level VARCHAR(30) NOT NULL DEFAULT 'READ',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    approved_by UUID REFERENCES iam_users(id),
    CONSTRAINT chk_permission_data_rights_subject CHECK (subject_type IN ('USER','GROUP','ROLE')),
    CONSTRAINT chk_permission_data_rights_access CHECK (access_level IN ('READ','WRITE','APPROVE','ADMIN')),
    CONSTRAINT chk_permission_data_rights_target CHECK (
        fund_id IS NOT NULL OR contract_id IS NOT NULL OR portfolio_id IS NOT NULL
    )
);
