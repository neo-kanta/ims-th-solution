DROP TRIGGER IF EXISTS trg_notification_settings_updated_at ON notification_settings;
DROP TABLE IF EXISTS notification_settings;

DROP TRIGGER IF EXISTS trg_audit_logs_no_delete ON audit_logs;
DROP TRIGGER IF EXISTS trg_audit_logs_no_update ON audit_logs;
DROP FUNCTION IF EXISTS prevent_audit_logs_modification();
DROP TABLE IF EXISTS audit_logs;

DROP TRIGGER IF EXISTS trg_permission_data_rights_updated_at ON permission_data_rights;
DROP TABLE IF EXISTS permission_data_rights;

DROP TRIGGER IF EXISTS trg_permission_function_rights_updated_at ON permission_function_rights;
DROP TABLE IF EXISTS permission_function_rights;
DROP TABLE IF EXISTS permission_function_definitions;

DROP TABLE IF EXISTS permission_change_request_labels;

DROP TRIGGER IF EXISTS trg_permission_labels_updated_at ON permission_labels;
DROP TABLE IF EXISTS permission_labels;

DROP TABLE IF EXISTS approval_workflow_events;
DROP TABLE IF EXISTS permission_request_revisions;
DROP TABLE IF EXISTS permission_request_checks;

DROP TRIGGER IF EXISTS trg_permission_request_comments_updated_at ON permission_request_comments;
DROP TABLE IF EXISTS permission_request_comments;

DROP TABLE IF EXISTS permission_request_step_approvers;

DROP TRIGGER IF EXISTS trg_permission_request_approval_steps_updated_at ON permission_request_approval_steps;
DROP TABLE IF EXISTS permission_request_approval_steps;

DROP TRIGGER IF EXISTS trg_approval_workflow_steps_updated_at ON approval_workflow_steps;
DROP TABLE IF EXISTS approval_workflow_steps;

DROP TRIGGER IF EXISTS trg_approval_workflow_settings_updated_at ON approval_workflow_settings;
DROP TABLE IF EXISTS approval_workflow_settings;

DROP TABLE IF EXISTS permission_change_items;
DROP SEQUENCE IF EXISTS permission_change_request_no_seq;

DROP TRIGGER IF EXISTS trg_permission_change_requests_updated_at ON permission_change_requests;
DROP TABLE IF EXISTS permission_change_requests;

DROP TABLE IF EXISTS permission_role_assignment_policies;

DROP TRIGGER IF EXISTS trg_permission_user_role_assignments_updated_at ON permission_user_role_assignments;
DROP TABLE IF EXISTS permission_user_role_assignments;

DROP TRIGGER IF EXISTS trg_permission_roles_updated_at ON permission_roles;
DROP TABLE IF EXISTS permission_roles;
