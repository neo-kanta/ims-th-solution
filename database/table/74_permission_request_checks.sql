-- Table: permission_request_checks
-- Source: 20260521000001_permissions__approval_workflow.up.sql
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
