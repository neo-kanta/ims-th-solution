// Approval API service. All calls go through the centralized useApi() helper
// (auth header + base URL). Request/response shapes are the generated types
// from ~/api/ims-api — nothing here is hand-typed.

import { useApi } from "~/composables/useApi";

import type {
  ActionInput,
  ApprovalEvent,
  ApprovalGroup,
  ApprovalGroupMember,
  ApprovalInboxList,
  ApprovalProcessConfig,
  ApprovalRequest,
  ApprovalRequestDetail,
  ApprovalRequestList,
  ApprovalSubjectStatus,
  ApprovalTeam,
  ApprovalTeamContract,
  ApprovalTeamMember,
  GroupInput,
  GroupMemberInput,
  ProcessConfigInput,
  ReorderInput,
  SubmitInput,
  TeamContractInput,
  TeamInput,
  TeamMemberInput,
} from "../types";

interface Envelope<T> {
  data: T;
  message?: string;
}

function unwrap<T>(e: Envelope<T>): T {
  return e.data;
}

function query(params: Record<string, string | number | undefined>): string {
  const sp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "") continue;
    sp.set(k, String(v));
  }
  const s = sp.toString();
  return s ? `?${s}` : "";
}

export const approvalApi = {
  // ── Runtime ────────────────────────────────────────────────────────────────
  async inbox(params: { status?: string; page?: number; limit?: number } = {}): Promise<ApprovalInboxList> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalInboxList>>(`/approvals/inbox${query(params)}`));
  },

  async requests(params: { status?: string; process_type?: string; page?: number; limit?: number } = {}): Promise<ApprovalRequestList> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalRequestList>>(`/approvals/requests${query(params)}`));
  },

  async request(id: string): Promise<ApprovalRequestDetail> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalRequestDetail>>(`/approvals/requests/${id}`));
  },

  async timeline(id: string): Promise<ApprovalEvent[]> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalEvent[]>>(`/approvals/requests/${id}/timeline`));
  },

  async subjectStatus(subjectType: string, subjectId: string): Promise<ApprovalSubjectStatus> {
    const { apiFetch } = useApi();
    return unwrap(
      await apiFetch<Envelope<ApprovalSubjectStatus>>(`/approvals/subjects/${subjectType}/${subjectId}/status`),
    );
  },

  async submit(body: SubmitInput): Promise<ApprovalRequest> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalRequest>>(`/approvals/submit`, { method: "POST", body }));
  },

  async approve(taskId: string, body: ActionInput): Promise<ApprovalRequest> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalRequest>>(`/approvals/tasks/${taskId}/approve`, { method: "POST", body }));
  },

  async reject(taskId: string, body: ActionInput): Promise<ApprovalRequest> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalRequest>>(`/approvals/tasks/${taskId}/reject`, { method: "POST", body }));
  },

  async withdraw(requestId: string): Promise<ApprovalRequest> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalRequest>>(`/approvals/requests/${requestId}/withdraw`, { method: "POST" }));
  },

  async cancel(requestId: string): Promise<ApprovalRequest> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalRequest>>(`/approvals/requests/${requestId}/cancel`, { method: "POST" }));
  },

  // ── Config: groups ───────────────────────────────────────────────────────────
  async listGroups(): Promise<ApprovalGroup[]> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalGroup[]>>(`/approval-config/groups`)) ?? [];
  },

  async createGroup(body: GroupInput): Promise<ApprovalGroup> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalGroup>>(`/approval-config/groups`, { method: "POST", body }));
  },

  async updateGroup(id: string, body: GroupInput): Promise<ApprovalGroup> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalGroup>>(`/approval-config/groups/${id}`, { method: "PUT", body }));
  },

  async listGroupMembers(id: string): Promise<ApprovalGroupMember[]> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalGroupMember[]>>(`/approval-config/groups/${id}/members`)) ?? [];
  },

  async addGroupMember(id: string, body: GroupMemberInput): Promise<ApprovalGroupMember> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalGroupMember>>(`/approval-config/groups/${id}/members`, { method: "POST", body }));
  },

  async updateGroupMember(id: string, memberId: string, body: GroupMemberInput): Promise<ApprovalGroupMember> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalGroupMember>>(`/approval-config/groups/${id}/members/${memberId}`, { method: "PUT", body }));
  },

  async approveGroupMember(id: string, memberId: string): Promise<ApprovalGroupMember> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalGroupMember>>(`/approval-config/groups/${id}/members/${memberId}/approve`, { method: "POST" }));
  },

  async revokeGroupMember(id: string, memberId: string): Promise<ApprovalGroupMember> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalGroupMember>>(`/approval-config/groups/${id}/members/${memberId}/revoke`, { method: "POST" }));
  },

  async reorderGroupMembers(id: string, body: ReorderInput): Promise<ApprovalGroupMember[]> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalGroupMember[]>>(`/approval-config/groups/${id}/members/reorder`, { method: "POST", body })) ?? [];
  },

  // ── Config: teams ────────────────────────────────────────────────────────────
  async listTeams(): Promise<ApprovalTeam[]> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalTeam[]>>(`/approval-config/teams`)) ?? [];
  },

  async createTeam(body: TeamInput): Promise<ApprovalTeam> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalTeam>>(`/approval-config/teams`, { method: "POST", body }));
  },

  async updateTeam(id: string, body: TeamInput): Promise<ApprovalTeam> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalTeam>>(`/approval-config/teams/${id}`, { method: "PUT", body }));
  },

  async assignTeamContract(id: string, body: TeamContractInput): Promise<ApprovalTeamContract> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalTeamContract>>(`/approval-config/teams/${id}/contracts`, { method: "POST", body }));
  },

  async listTeamMembers(id: string): Promise<ApprovalTeamMember[]> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalTeamMember[]>>(`/approval-config/teams/${id}/members`)) ?? [];
  },

  async addTeamMember(id: string, body: TeamMemberInput): Promise<ApprovalTeamMember> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalTeamMember>>(`/approval-config/teams/${id}/members`, { method: "POST", body }));
  },

  async updateTeamMember(id: string, memberId: string, body: TeamMemberInput): Promise<ApprovalTeamMember> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalTeamMember>>(`/approval-config/teams/${id}/members/${memberId}`, { method: "PUT", body }));
  },

  async removeTeamMember(id: string, memberId: string): Promise<void> {
    const { apiFetch } = useApi();
    await apiFetch<void>(`/approval-config/teams/${id}/members/${memberId}`, { method: "DELETE" });
  },

  // ── Config: processes ────────────────────────────────────────────────────────
  async listProcesses(params: { process_type?: string; active_only?: boolean } = {}): Promise<ApprovalProcessConfig[]> {
    const { apiFetch } = useApi();
    const q = query({ process_type: params.process_type, active_only: params.active_only ? "true" : undefined });
    return unwrap(await apiFetch<Envelope<ApprovalProcessConfig[]>>(`/approval-config/processes${q}`)) ?? [];
  },

  async getProcess(id: string): Promise<ApprovalProcessConfig> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalProcessConfig>>(`/approval-config/processes/${id}`));
  },

  async createProcess(body: ProcessConfigInput): Promise<ApprovalProcessConfig> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalProcessConfig>>(`/approval-config/processes`, { method: "POST", body }));
  },

  async updateProcess(id: string, body: ProcessConfigInput): Promise<ApprovalProcessConfig> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalProcessConfig>>(`/approval-config/processes/${id}`, { method: "PUT", body }));
  },

  async activateProcess(id: string): Promise<ApprovalProcessConfig> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalProcessConfig>>(`/approval-config/processes/${id}/activate`, { method: "POST" }));
  },

  async deactivateProcess(id: string): Promise<ApprovalProcessConfig> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<ApprovalProcessConfig>>(`/approval-config/processes/${id}/deactivate`, { method: "POST" }));
  },
};

/** Extracts a human-readable error message from a fetch error. */
export function approvalErrorMessage(err: unknown, fallback: string): string {
  if (err && typeof err === "object") {
    const data = (err as { data?: { error?: unknown } }).data;
    if (data && typeof data.error === "string" && data.error.trim()) return data.error;
    const status = (err as { status?: number; statusCode?: number }).status ?? (err as { statusCode?: number }).statusCode;
    if (status === 403) return "You do not have permission to perform this action.";
    const message = (err as { message?: unknown }).message;
    if (typeof message === "string" && message.trim()) return message;
  }
  return fallback;
}

/** Returns the HTTP status from a fetch error, if available. */
export function approvalErrorStatus(err: unknown): number | null {
  if (err && typeof err === "object") {
    const s = (err as { status?: number; statusCode?: number }).status ?? (err as { statusCode?: number }).statusCode;
    if (typeof s === "number") return s;
  }
  return null;
}
