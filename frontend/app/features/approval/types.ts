// Approval feature types.
//
// Per project policy, API shapes are NOT hand-written: every type below is a
// direct alias of a schema generated into `~/api/ims-api` from the backend
// Swagger. UI-only helper types live at the bottom.

import type { components } from "~/api/ims-api";

type Schemas = components["schemas"];

// ── Responses ────────────────────────────────────────────────────────────────
export type ApprovalGroup = Schemas["GroupResponse"];
export type ApprovalGroupMember = Schemas["GroupMemberResponse"];
export type ApprovalTeam = Schemas["TeamResponse"];
export type ApprovalTeamContract = Schemas["TeamContractResponse"];
export type ApprovalTeamMember = Schemas["TeamMemberResponse"];
export type ApprovalProcessConfig = Schemas["ProcessConfigResponse"];
export type ApprovalProcessStage = Schemas["StageResponse"];
export type ApprovalRequest = Schemas["RequestResponse"];
export type ApprovalTask = Schemas["TaskResponse"];
export type ApprovalEvent = Schemas["EventResponse"];
export type ApprovalSignature = Schemas["SignatureResponse"];
export type ApprovalRequestDetail = Schemas["RequestDetailResponse"];
export type ApprovalInboxItem = Schemas["InboxItemResponse"];
export type ApprovalInboxList = Schemas["InboxListResponse"];
export type ApprovalRequestList = Schemas["RequestListResponse"];
export type ApprovalSubjectStatus = Schemas["SubjectStatusResponse"];

// ── Request bodies ───────────────────────────────────────────────────────────
export type GroupInput = Schemas["GroupRequest"];
export type GroupMemberInput = Schemas["GroupMemberRequest"];
export type ReorderInput = Schemas["ReorderRequest"];
export type TeamInput = Schemas["TeamRequest"];
export type TeamContractInput = Schemas["TeamContractRequest"];
export type TeamMemberInput = Schemas["TeamMemberRequest"];
export type StageInput = Schemas["StageRequest"];
export type ProcessConfigInput = Schemas["ProcessConfigRequest"];
export type SubmitInput = Schemas["SubmitRequest"];
export type ActionInput = Schemas["ActionRequest"];

// ── Enumerations (stable contracts shared with the backend) ──────────────────
export const PROCESS_TYPES = [
  "INVESTMENT_ANALYSIS_REPORT",
  "INVESTMENT_DECISION",
  "INVESTMENT_CANCELLATION",
  "WORKFLOW_OPERATION",
  "LEAVE_REQUEST",
  "LEAVE_CANCELLATION",
  "DELEGATION_REQUEST",
] as const;
export type ProcessType = (typeof PROCESS_TYPES)[number];

export const CONTRACT_TYPES = ["FUND", "DISCRETIONARY", "COMPANY"] as const;
export type ContractType = (typeof CONTRACT_TYPES)[number];

export const APPROVER_MODES = [
  "SINGLE_USER",
  "GROUP_PRIORITY",
  "GROUP_ANY",
  "TEAM_MINIMUM",
] as const;
export type ApproverMode = (typeof APPROVER_MODES)[number];

export const REQUEST_STATUSES = [
  "DRAFT",
  "SUBMITTED",
  "PENDING_APPROVAL",
  "APPROVED",
  "REJECTED",
  "CANCELLED",
  "WITHDRAWN",
  "REVOKED",
] as const;
export type RequestStatus = (typeof REQUEST_STATUSES)[number];

export const GROUP_MEMBER_TYPES = ["MEMBER", "SUPERVISOR"] as const;
export const GROUP_MEMBER_STATUSES = ["PENDING", "APPROVED", "REVOKED"] as const;
export const TEAM_MEMBER_TYPES = ["ORDER_SUBMITTER", "REVIEWER_AGENT"] as const;

// ── UI helper types ──────────────────────────────────────────────────────────
export interface ApprovalListState<T> {
  items: T[];
  loading: boolean;
  error: string | null;
}

export interface ApprovalTarget {
  moduleCode: string;
  processType: string;
  recordType: string;
  recordId: string;
  businessDate?: string;
  title?: string;
  description?: string;
}

export type ApprovalAction = "SUBMIT" | "CANCEL_SUBMISSION" | "APPROVE" | "REJECT" | "REVOKE_APPROVAL";

export type ApprovalStatus =
  | "DRAFT"
  | "NOT_SUBMITTED"
  | "PENDING_APPROVAL"
  | "APPROVED"
  | "REJECTED"
  | "CANCELLED"
  | "REVOKED";

export interface ApprovalApprover {
  userId: string;
  displayName: string;
  roleName?: string;
  priority?: number;
  status: "WAITING" | "APPROVED" | "REJECTED" | "SKIPPED";
  actedAt?: string;
  isAgent?: boolean;
  principalUserName?: string;
  remark?: string;
}

export interface ApprovalStage {
  stageNo: number;
  stageName: string;
  approvalMode: "SINGLE" | "GROUP_ANY" | "TEAM_STAMP";
  status: "PENDING" | "APPROVED" | "REJECTED" | "SKIPPED";
  requiredStampCount?: number;
  currentStampCount?: number;
  approvers: ApprovalApprover[];
}

export interface ApprovalInstance {
  approvalId: string;
  target: ApprovalTarget;
  status: ApprovalStatus;
  currentStageNo?: number;
  nextApproverNames: string[];
  submittedBy?: string;
  submittedAt?: string;
  completedAt?: string;
  stages: ApprovalStage[];
  history: ApprovalHistoryItem[];
}

export interface ApprovalHistoryItem {
  id: string;
  action: ApprovalAction;
  actionBy: string;
  actionByDisplayName: string;
  actedAt: string;
  stageNo?: number;
  remark?: string;
  isAgent?: boolean;
  principalUserName?: string;
}

