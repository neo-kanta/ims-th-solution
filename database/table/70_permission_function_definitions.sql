-- Table: permission_function_definitions
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_function_definitions (
    code VARCHAR(140) PRIMARY KEY,
    module VARCHAR(80) NOT NULL,
    screen VARCHAR(100),
    action VARCHAR(60) NOT NULL,
    name VARCHAR(180) NOT NULL,
    description TEXT,
    deprecated_at TIMESTAMPTZ
);
