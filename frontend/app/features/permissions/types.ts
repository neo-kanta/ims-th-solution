export interface PermissionLabel {
  id: string;
  label_code: string;
  label_name: string;
  label_type: string;
  color: string;
  description: string;
  is_system: boolean;
  is_active: boolean;
}

export interface PermissionChangeItem {
  ID?: string;
  id?: string;
  item_type: string;
  target_table: string;
  target_id: string;
  action_type: string;
  before_json: Record<string, unknown>;
  after_json: Record<string, unknown>;
  created_at?: string;
}

export interface PermissionCheck {
  id: string;
  check_code: string;
  check_name: string;
  status: "PASSED" | "WARNING" | "FAILED";
  severity: "INFO" | "WARNING" | "BLOCKER";
  message: string;
  created_at: string;
}

export interface PermissionStepApprover {
  id: string;
  approver_user_id?: string | null;
  approver_display_name?: string;
  approver_role_code: string;
  approval_status: "PENDING" | "APPROVED" | "CHANGES_REQUESTED" | "REJECTED";
  approval_comment: string;
  approved_at?: string | null;
}

export interface PermissionApprovalStep {
  id: string;
  step_no: number;
  step_name: string;
  approval_mode: string;
  status: "NOT_STARTED" | "PENDING" | "APPROVED" | "CHANGES_REQUESTED" | "REJECTED" | "SKIPPED";
  min_approvals_required: number;
  approvals_received: number;
  started_at?: string | null;
  completed_at?: string | null;
  approvers: PermissionStepApprover[];
}

export interface PermissionComment {
  id: string;
  user_name: string;
  comment: string;
  created_at: string;
}

export interface PermissionWorkflowEvent {
  id: string;
  actor_name: string;
  event_type: string;
  comment: string;
  created_at: string;
}

export interface PermissionChangeRequest {
  id: string;
  request_no: string;
  title: string;
  description: string;
  request_type: string;
  status: "DRAFT" | "READY_FOR_REVIEW" | "CHANGES_REQUESTED" | "APPROVED" | "REJECTED" | "MERGED" | "CLOSED" | "CANCELLED";
  risk_level: "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";
  target_entity_type: string;
  target_entity_id: string;
  created_by: string;
  created_by_name: string;
  created_at: string;
  updated_at: string;
  items: PermissionChangeItem[];
  steps: PermissionApprovalStep[];
  comments: PermissionComment[];
  checks: PermissionCheck[];
  labels: PermissionLabel[];
  events: PermissionWorkflowEvent[];
}

export interface PermissionRole {
  id: string;
  role_code: string;
  role_name: string;
  department: string;
  role_category: string;
  priority_rank: number;
  assignment_scope: string;
  can_request_role_assignment: boolean;
  can_approve_role_assignment: boolean;
  is_high_risk: boolean;
  is_active: boolean;
  description: string;
}

export interface PermissionUserSummary {
  id: string;
  username: string;
  display_name: string;
  email: string;
  is_active: boolean;
  is_locked: boolean;
  groups: string[];
  roles: string[];
}

export interface PermissionGroupSummary {
  id: string;
  name: string;
  description: string;
  is_active: boolean;
  members_count: number;
}

export interface EffectivePermissions {
  user_id: string;
  direct_function_permissions: unknown[];
  group_function_permissions: unknown[];
  role_function_permissions: unknown[];
  direct_data_permissions: unknown[];
  group_data_permissions: unknown[];
  role_data_permissions: unknown[];
  final_function_permissions: string[];
  final_contract_permissions: string[];
  roles: PermissionRole[];
}
