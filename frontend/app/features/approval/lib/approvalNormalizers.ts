import type { ApprovalInboxItem, ApprovalRequest, ApprovalTask } from "../types";

export interface NormalizedRequest {
  id: string;
  requestNumber: string;
  module: string;
  processType: string;
  recordType: string;
  recordTitle: string;
  submitter: string;
  currentStage: number | string;
  nextApprover: string;
  submittedAt: string;
  status: string;
}

export function normalizeInboxItem(item: ApprovalInboxItem): NormalizedRequest {
  const req = (item.request ?? {}) as Partial<ApprovalRequest>;
  const task = (item.task ?? {}) as Partial<ApprovalTask>;
  return {
    id: req.id ?? "",
    requestNumber: req.request_number ?? "",
    module: req.contract_type || req.process_type || "SYSTEM",
    processType: req.process_type ?? "",
    recordType: req.subject_type ?? "",
    recordTitle: req.subject?.display_label || req.subject_title || req.subject_reference || "—",
    submitter: req.submitter?.display_name || req.submitter_name || "—",
    currentStage: req.current_stage_number ?? 1,
    nextApprover: task.assigned_user?.display_name || task.assigned_user_name || "—",
    submittedAt: req.submitted_at ?? "",
    status: req.status ?? "DRAFT",
  };
}

export function normalizeRequestItem(req: ApprovalRequest): NormalizedRequest {
  return {
    id: req.id ?? "",
    requestNumber: req.request_number ?? "",
    module: req.contract_type || "SYSTEM",
    processType: req.process_type ?? "",
    recordType: req.subject_type ?? "",
    recordTitle: req.subject?.display_label || req.subject_title || req.subject_reference || "—",
    submitter: req.submitter?.display_name || req.submitter_name || "—",
    currentStage: req.current_stage_number ?? "—",
    nextApprover: req.status === "PENDING_APPROVAL" ? "Assigned Approver(s)" : "—",
    submittedAt: req.submitted_at ?? "",
    status: req.status ?? "DRAFT",
  };
}
