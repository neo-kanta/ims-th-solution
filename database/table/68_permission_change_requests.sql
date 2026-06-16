-- Table: permission_change_requests
-- Source: 20260521000001_permissions__approval_workflow.up.sql
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
