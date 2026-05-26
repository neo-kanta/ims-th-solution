-- =============================================================================
-- Permission Management + Approval Workflow
-- =============================================================================
-- This migration keeps the existing `permissions_*` grant tables intact and
-- adds the financial-grade permission request workflow surface requested by
-- the permission-management module. Existing IAM permission evaluation can
-- continue to read legacy grants while the new module records auditable,
-- step-based permission changes.
-- =============================================================================

CREATE TABLE permission_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_code VARCHAR(80) NOT NULL,
    role_name VARCHAR(160) NOT NULL,
    department VARCHAR(120),
    role_category VARCHAR(80),
    priority_rank INTEGER NOT NULL,
    assignment_scope VARCHAR(80) NOT NULL,
    can_request_role_assignment BOOLEAN NOT NULL DEFAULT false,
    can_approve_role_assignment BOOLEAN NOT NULL DEFAULT false,
    is_high_risk BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_roles_code UNIQUE (role_code),
    CONSTRAINT chk_permission_roles_priority CHECK (priority_rank >= 0)
);

CREATE TRIGGER trg_permission_roles_updated_at
    BEFORE UPDATE ON permission_roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE permission_user_role_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES permission_roles(id) ON DELETE RESTRICT,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    assigned_by UUID REFERENCES iam_users(id),
    approved_by UUID REFERENCES iam_users(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    effective_from TIMESTAMPTZ,
    effective_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_permission_user_role_assignments_status
        CHECK (status IN ('PENDING','APPROVED','REJECTED','REVOKED','EXPIRED')),
    CONSTRAINT chk_permission_user_role_assignments_effective_window
        CHECK (effective_to IS NULL OR effective_from IS NULL OR effective_to > effective_from)
);

CREATE UNIQUE INDEX uq_permission_user_role_active
    ON permission_user_role_assignments (user_id, role_id)
    WHERE status IN ('PENDING','APPROVED') AND effective_to IS NULL;

CREATE INDEX idx_permission_user_role_user
    ON permission_user_role_assignments (user_id, status);

CREATE TRIGGER trg_permission_user_role_assignments_updated_at
    BEFORE UPDATE ON permission_user_role_assignments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE permission_role_assignment_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grantor_role_id UUID NOT NULL REFERENCES permission_roles(id) ON DELETE CASCADE,
    max_target_priority_exclusive INTEGER NOT NULL,
    assignment_scope VARCHAR(80) NOT NULL,
    min_approvals_required INTEGER NOT NULL DEFAULT 1,
    can_request BOOLEAN NOT NULL DEFAULT true,
    can_direct_merge BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_permission_role_assignment_policies_priority
        CHECK (max_target_priority_exclusive >= 0),
    CONSTRAINT chk_permission_role_assignment_policies_approvals
        CHECK (min_approvals_required >= 1)
);

CREATE TABLE permission_change_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_no VARCHAR(40) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    request_type VARCHAR(60) NOT NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'DRAFT',
    risk_level VARCHAR(20) NOT NULL DEFAULT 'LOW',
    target_entity_type VARCHAR(60),
    target_entity_id VARCHAR(120),
    created_by UUID NOT NULL REFERENCES iam_users(id),
    assigned_to UUID REFERENCES iam_users(id),
    submitted_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ,
    approved_by UUID REFERENCES iam_users(id),
    merged_at TIMESTAMPTZ,
    merged_by UUID REFERENCES iam_users(id),
    rejected_at TIMESTAMPTZ,
    rejected_by UUID REFERENCES iam_users(id),
    rejection_reason TEXT,
    closed_at TIMESTAMPTZ,
    closed_by UUID REFERENCES iam_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_change_requests_no UNIQUE (request_no),
    CONSTRAINT chk_permission_change_requests_status CHECK (status IN (
        'DRAFT','READY_FOR_REVIEW','CHANGES_REQUESTED','APPROVED','REJECTED',
        'MERGED','CLOSED','CANCELLED'
    )),
    CONSTRAINT chk_permission_change_requests_risk CHECK (risk_level IN (
        'LOW','MEDIUM','HIGH','CRITICAL'
    ))
);

CREATE INDEX idx_permission_change_requests_status
    ON permission_change_requests (status, updated_at DESC);

CREATE INDEX idx_permission_change_requests_created_by
    ON permission_change_requests (created_by, created_at DESC);

CREATE TRIGGER trg_permission_change_requests_updated_at
    BEFORE UPDATE ON permission_change_requests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE SEQUENCE permission_change_request_no_seq START WITH 1 INCREMENT BY 1;

CREATE TABLE permission_change_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    item_type VARCHAR(60) NOT NULL,
    target_table VARCHAR(100),
    target_id VARCHAR(120),
    action_type VARCHAR(60) NOT NULL,
    before_json JSONB NOT NULL DEFAULT '{}',
    after_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_permission_change_items_request
    ON permission_change_items (request_id, created_at);

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

CREATE INDEX idx_approval_workflow_settings_lookup
    ON approval_workflow_settings (module, request_type, risk_level)
    WHERE is_active = true;

CREATE TRIGGER trg_approval_workflow_settings_updated_at
    BEFORE UPDATE ON approval_workflow_settings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE approval_workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_setting_id UUID NOT NULL REFERENCES approval_workflow_settings(id) ON DELETE CASCADE,
    step_no INTEGER NOT NULL,
    step_name VARCHAR(180) NOT NULL,
    step_description TEXT,
    approval_mode VARCHAR(40) NOT NULL DEFAULT 'SINGLE',
    approver_type VARCHAR(40) NOT NULL DEFAULT 'ROLE',
    required_role_code VARCHAR(80),
    required_group_id UUID REFERENCES permissions_groups(id),
    required_user_id UUID REFERENCES iam_users(id),
    min_approvals_required INTEGER NOT NULL DEFAULT 1,
    allow_delegation BOOLEAN NOT NULL DEFAULT false,
    is_required BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_workflow_steps_setting_step UNIQUE (workflow_setting_id, step_no),
    CONSTRAINT chk_approval_workflow_steps_mode CHECK (approval_mode IN (
        'SINGLE','ANY_OF_GROUP','ALL_OF_GROUP','QUORUM','ROLE_BASED','USER_BASED'
    )),
    CONSTRAINT chk_approval_workflow_steps_approver_type CHECK (approver_type IN (
        'ROLE','GROUP','USER','REQUEST_TARGET_OWNER'
    )),
    CONSTRAINT chk_approval_workflow_steps_min CHECK (min_approvals_required >= 1)
);

CREATE TRIGGER trg_approval_workflow_steps_updated_at
    BEFORE UPDATE ON approval_workflow_steps
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE permission_request_approval_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    workflow_setting_id UUID NOT NULL REFERENCES approval_workflow_settings(id) ON DELETE RESTRICT,
    step_no INTEGER NOT NULL,
    step_name VARCHAR(180) NOT NULL,
    approval_mode VARCHAR(40) NOT NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'NOT_STARTED',
    min_approvals_required INTEGER NOT NULL DEFAULT 1,
    approvals_received INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_request_approval_steps_request_step UNIQUE (request_id, step_no),
    CONSTRAINT chk_permission_request_approval_steps_status CHECK (status IN (
        'NOT_STARTED','PENDING','APPROVED','CHANGES_REQUESTED','REJECTED','SKIPPED'
    )),
    CONSTRAINT chk_permission_request_approval_steps_counts CHECK (
        min_approvals_required >= 1 AND approvals_received >= 0
    )
);

CREATE INDEX idx_permission_request_approval_steps_request
    ON permission_request_approval_steps (request_id, step_no);

CREATE TRIGGER trg_permission_request_approval_steps_updated_at
    BEFORE UPDATE ON permission_request_approval_steps
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE permission_request_step_approvers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    approval_step_id UUID NOT NULL REFERENCES permission_request_approval_steps(id) ON DELETE CASCADE,
    approver_user_id UUID REFERENCES iam_users(id),
    approver_role_code VARCHAR(80),
    approval_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
    approval_comment TEXT,
    is_delegated BOOLEAN NOT NULL DEFAULT false,
    delegated_from_user_id UUID REFERENCES iam_users(id),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_permission_request_step_approvers_status CHECK (
        approval_status IN ('PENDING','APPROVED','CHANGES_REQUESTED','REJECTED')
    )
);

CREATE INDEX idx_permission_request_step_approvers_step
    ON permission_request_step_approvers (approval_step_id, approval_status);

CREATE INDEX idx_permission_request_step_approvers_user
    ON permission_request_step_approvers (approver_user_id, approval_status);

CREATE TABLE permission_request_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES iam_users(id),
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_permission_request_comments_request
    ON permission_request_comments (request_id, created_at)
    WHERE deleted_at IS NULL;

CREATE TRIGGER trg_permission_request_comments_updated_at
    BEFORE UPDATE ON permission_request_comments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE permission_request_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    check_code VARCHAR(100) NOT NULL,
    check_name VARCHAR(180) NOT NULL,
    status VARCHAR(20) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'INFO',
    message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_permission_request_checks_status CHECK (status IN ('PASSED','WARNING','FAILED')),
    CONSTRAINT chk_permission_request_checks_severity CHECK (severity IN ('INFO','WARNING','BLOCKER'))
);

CREATE INDEX idx_permission_request_checks_request
    ON permission_request_checks (request_id, status, severity);

CREATE TABLE permission_request_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    revision_no INTEGER NOT NULL,
    changed_by UUID NOT NULL REFERENCES iam_users(id),
    change_summary TEXT,
    snapshot_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_request_revisions_request_revision UNIQUE (request_id, revision_no)
);

CREATE TABLE approval_workflow_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    approval_step_id UUID REFERENCES permission_request_approval_steps(id) ON DELETE SET NULL,
    actor_user_id UUID REFERENCES iam_users(id),
    event_type VARCHAR(80) NOT NULL,
    before_json JSONB NOT NULL DEFAULT '{}',
    after_json JSONB NOT NULL DEFAULT '{}',
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_approval_workflow_events_request
    ON approval_workflow_events (request_id, created_at);

CREATE TABLE permission_labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label_code VARCHAR(100) NOT NULL,
    label_name VARCHAR(140) NOT NULL,
    label_type VARCHAR(40) NOT NULL,
    color VARCHAR(20) NOT NULL DEFAULT '#6e7781',
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_labels_code UNIQUE (label_code)
);

CREATE INDEX idx_permission_labels_type
    ON permission_labels (label_type, is_active);

CREATE TRIGGER trg_permission_labels_updated_at
    BEFORE UPDATE ON permission_labels
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE permission_change_request_labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES permission_labels(id) ON DELETE CASCADE,
    added_by UUID REFERENCES iam_users(id),
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_change_request_labels UNIQUE (request_id, label_id)
);

CREATE TABLE permission_function_definitions (
    code VARCHAR(140) PRIMARY KEY,
    module VARCHAR(80) NOT NULL,
    screen VARCHAR(100),
    action VARCHAR(60) NOT NULL,
    name VARCHAR(180) NOT NULL,
    description TEXT,
    deprecated_at TIMESTAMPTZ
);

CREATE INDEX idx_permission_function_definitions_module
    ON permission_function_definitions (module, screen)
    WHERE deprecated_at IS NULL;

CREATE TABLE permission_function_rights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_type VARCHAR(20) NOT NULL,
    subject_id UUID NOT NULL,
    permission_code VARCHAR(140) NOT NULL REFERENCES permission_function_definitions(code) ON UPDATE CASCADE ON DELETE RESTRICT,
    can_view BOOLEAN NOT NULL DEFAULT false,
    can_search BOOLEAN NOT NULL DEFAULT false,
    can_add BOOLEAN NOT NULL DEFAULT false,
    can_edit BOOLEAN NOT NULL DEFAULT false,
    can_delete BOOLEAN NOT NULL DEFAULT false,
    can_approve BOOLEAN NOT NULL DEFAULT false,
    can_revoke_approval BOOLEAN NOT NULL DEFAULT false,
    can_export BOOLEAN NOT NULL DEFAULT false,
    can_configure BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    approved_by UUID REFERENCES iam_users(id),
    CONSTRAINT uq_permission_function_rights_subject_code UNIQUE (subject_type, subject_id, permission_code),
    CONSTRAINT chk_permission_function_rights_subject CHECK (subject_type IN ('USER','GROUP','ROLE'))
);

CREATE INDEX idx_permission_function_rights_subject
    ON permission_function_rights (subject_type, subject_id);

CREATE TRIGGER trg_permission_function_rights_updated_at
    BEFORE UPDATE ON permission_function_rights
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

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

CREATE UNIQUE INDEX uq_permission_data_rights_subject_target
    ON permission_data_rights (
        subject_type,
        subject_id,
        COALESCE(fund_id, '00000000-0000-0000-0000-000000000000'::uuid),
        COALESCE(contract_id, '00000000-0000-0000-0000-000000000000'::uuid),
        COALESCE(portfolio_id, '00000000-0000-0000-0000-000000000000'::uuid),
        access_level
    );

CREATE INDEX idx_permission_data_rights_subject
    ON permission_data_rights (subject_type, subject_id);

CREATE TRIGGER trg_permission_data_rights_updated_at
    BEFORE UPDATE ON permission_data_rights
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID REFERENCES iam_users(id),
    action VARCHAR(120) NOT NULL,
    module VARCHAR(80) NOT NULL,
    entity_type VARCHAR(120),
    entity_id VARCHAR(160),
    before_json JSONB NOT NULL DEFAULT '{}',
    after_json JSONB NOT NULL DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    correlation_id VARCHAR(160),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_actor
    ON audit_logs (actor_user_id, created_at DESC);

CREATE INDEX idx_audit_logs_entity
    ON audit_logs (entity_type, entity_id, created_at DESC);

CREATE INDEX idx_audit_logs_module_action
    ON audit_logs (module, action, created_at DESC);

CREATE OR REPLACE FUNCTION prevent_audit_logs_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_logs are immutable and cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_audit_logs_no_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_logs_modification();

CREATE TRIGGER trg_audit_logs_no_delete
    BEFORE DELETE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_logs_modification();

CREATE TABLE notification_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_code VARCHAR(100) NOT NULL,
    channel VARCHAR(40) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    target_scope VARCHAR(80) NOT NULL DEFAULT 'APPROVERS',
    template_subject VARCHAR(255) NOT NULL DEFAULT '',
    template_body TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_notification_settings_event_channel UNIQUE (event_code, channel)
);

CREATE TRIGGER trg_notification_settings_updated_at
    BEFORE UPDATE ON notification_settings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
