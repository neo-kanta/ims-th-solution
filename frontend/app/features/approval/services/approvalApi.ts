// Approval API service. All calls go through the centralized useOpenApiClient() helper
// (auth header + base URL). Request/response shapes are the generated types
// from ~/api/ims-api — nothing here is hand-typed.

import { useOpenApiClient, unwrapOpenApiResponse } from "~/api/openapi";

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

export const approvalApi = {
  // ── Runtime ────────────────────────────────────────────────────────────────
  async inbox(params: { status?: string; page?: number; limit?: number } = {}): Promise<ApprovalInboxList> {
    const client = useOpenApiClient();
    const result = await client.GET("/approvals/inbox", {
      params: { query: params },
    });
    return unwrapOpenApiResponse(result);
  },

  async requests(params: { status?: string; process_type?: string; page?: number; limit?: number } = {}): Promise<ApprovalRequestList> {
    const client = useOpenApiClient();
    const result = await client.GET("/approvals/requests", {
      params: { query: params },
    });
    return unwrapOpenApiResponse(result);
  },

  async request(requestId: string): Promise<ApprovalRequestDetail> {
    const client = useOpenApiClient();
    const result = await client.GET("/approvals/requests/{requestId}", {
      params: { path: { requestId } },
    });
    return unwrapOpenApiResponse(result);
  },

  async timeline(requestId: string): Promise<ApprovalEvent[]> {
    const client = useOpenApiClient();
    const result = await client.GET("/approvals/requests/{requestId}/timeline", {
      params: { path: { requestId } },
    });
    return unwrapOpenApiResponse(result);
  },

  async subjectStatus(subjectType: string, subjectId: string): Promise<ApprovalSubjectStatus> {
    const client = useOpenApiClient();
    const result = await client.GET("/approvals/subjects/{subjectType}/{subjectId}/status", {
      params: { path: { subjectType, subjectId } },
    });
    return unwrapOpenApiResponse(result);
  },

  async submit(body: SubmitInput): Promise<ApprovalRequest> {
    const client = useOpenApiClient();
    const result = await client.POST("/approvals/submit", { body });
    return unwrapOpenApiResponse(result);
  },

  async approve(taskId: string, body: ActionInput): Promise<ApprovalRequest> {
    const client = useOpenApiClient();
    const result = await client.POST("/approvals/tasks/{taskId}/approve", {
      params: { path: { taskId } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  async reject(taskId: string, body: ActionInput): Promise<ApprovalRequest> {
    const client = useOpenApiClient();
    const result = await client.POST("/approvals/tasks/{taskId}/reject", {
      params: { path: { taskId } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  async withdraw(requestId: string): Promise<ApprovalRequest> {
    const client = useOpenApiClient();
    const result = await client.POST("/approvals/requests/{requestId}/withdraw", {
      params: { path: { requestId } },
    });
    return unwrapOpenApiResponse(result);
  },

  async cancel(requestId: string): Promise<ApprovalRequest> {
    const client = useOpenApiClient();
    const result = await client.POST("/approvals/requests/{requestId}/cancel", {
      params: { path: { requestId } },
    });
    return unwrapOpenApiResponse(result);
  },

  async revoke(requestId: string, body: ActionInput): Promise<ApprovalRequest> {
    const client = useOpenApiClient();
    const result = await client.POST("/approvals/requests/{requestId}/revoke", {
      params: { path: { requestId } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  // ── Config: groups ───────────────────────────────────────────────────────────
  async listGroups(): Promise<ApprovalGroup[]> {
    const client = useOpenApiClient();
    const result = await client.GET("/approval-config/groups");
    return unwrapOpenApiResponse(result) ?? [];
  },

  async createGroup(body: GroupInput): Promise<ApprovalGroup> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/groups", { body });
    return unwrapOpenApiResponse(result);
  },

  async updateGroup(id: string, body: GroupInput): Promise<ApprovalGroup> {
    const client = useOpenApiClient();
    const result = await client.PUT("/approval-config/groups/{id}", {
      params: { path: { id } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  async listGroupMembers(id: string): Promise<ApprovalGroupMember[]> {
    const client = useOpenApiClient();
    const result = await client.GET("/approval-config/groups/{id}/members", {
      params: { path: { id } },
    });
    return unwrapOpenApiResponse(result) ?? [];
  },

  async addGroupMember(id: string, body: GroupMemberInput): Promise<ApprovalGroupMember> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/groups/{id}/members", {
      params: { path: { id } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  async updateGroupMember(id: string, memberId: string, body: GroupMemberInput): Promise<ApprovalGroupMember> {
    const client = useOpenApiClient();
    const result = await client.PUT("/approval-config/groups/{id}/members/{memberId}", {
      params: { path: { id, memberId } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  async approveGroupMember(id: string, memberId: string): Promise<ApprovalGroupMember> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/groups/{id}/members/{memberId}/approve", {
      params: { path: { id, memberId } },
    });
    return unwrapOpenApiResponse(result);
  },

  async revokeGroupMember(id: string, memberId: string): Promise<ApprovalGroupMember> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/groups/{id}/members/{memberId}/revoke", {
      params: { path: { id, memberId } },
    });
    return unwrapOpenApiResponse(result);
  },

  async reorderGroupMembers(id: string, body: ReorderInput): Promise<ApprovalGroupMember[]> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/groups/{id}/members/reorder", {
      params: { path: { id } },
      body,
    });
    return unwrapOpenApiResponse(result) ?? [];
  },

  // ── Config: teams ────────────────────────────────────────────────────────────
  async listTeams(): Promise<ApprovalTeam[]> {
    const client = useOpenApiClient();
    const result = await client.GET("/approval-config/teams");
    return unwrapOpenApiResponse(result) ?? [];
  },

  async createTeam(body: TeamInput): Promise<ApprovalTeam> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/teams", { body });
    return unwrapOpenApiResponse(result);
  },

  async updateTeam(id: string, body: TeamInput): Promise<ApprovalTeam> {
    const client = useOpenApiClient();
    const result = await client.PUT("/approval-config/teams/{id}", {
      params: { path: { id } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  async listTeamContracts(id: string): Promise<ApprovalTeamContract[]> {
    const client = useOpenApiClient();
    const result = await client.GET("/approval-config/teams/{id}/contracts", {
      params: { path: { id } },
    });
    return unwrapOpenApiResponse(result) ?? [];
  },

  async assignTeamContract(id: string, body: TeamContractInput): Promise<ApprovalTeamContract> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/teams/{id}/contracts", {
      params: { path: { id } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  async listTeamMembers(id: string): Promise<ApprovalTeamMember[]> {
    const client = useOpenApiClient();
    const result = await client.GET("/approval-config/teams/{id}/members", {
      params: { path: { id } },
    });
    return unwrapOpenApiResponse(result) ?? [];
  },

  async addTeamMember(id: string, body: TeamMemberInput): Promise<ApprovalTeamMember> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/teams/{id}/members", {
      params: { path: { id } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  async updateTeamMember(id: string, memberId: string, body: TeamMemberInput): Promise<ApprovalTeamMember> {
    const client = useOpenApiClient();
    const result = await client.PUT("/approval-config/teams/{id}/members/{memberId}", {
      params: { path: { id, memberId } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  async removeTeamMember(id: string, memberId: string): Promise<void> {
    const client = useOpenApiClient();
    const result = await client.DELETE("/approval-config/teams/{id}/members/{memberId}", {
      params: { path: { id, memberId } },
    });
    unwrapOpenApiResponse(result);
  },

  // ── Config: processes ────────────────────────────────────────────────────────
  async listProcesses(params: { process_type?: string; active_only?: boolean } = {}): Promise<ApprovalProcessConfig[]> {
    const client = useOpenApiClient();
    const query: { process_type?: string; active_only?: boolean } = {};
    if (params.process_type) query.process_type = params.process_type;
    if (params.active_only !== undefined) query.active_only = params.active_only;

    const result = await client.GET("/approval-config/processes", {
      params: { query },
    });
    return unwrapOpenApiResponse(result) ?? [];
  },

  async getProcess(id: string): Promise<ApprovalProcessConfig> {
    const client = useOpenApiClient();
    const result = await client.GET("/approval-config/processes/{id}", {
      params: { path: { id } },
    });
    return unwrapOpenApiResponse(result);
  },

  async createProcess(body: ProcessConfigInput): Promise<ApprovalProcessConfig> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/processes", { body });
    return unwrapOpenApiResponse(result);
  },

  async updateProcess(id: string, body: ProcessConfigInput): Promise<ApprovalProcessConfig> {
    const client = useOpenApiClient();
    const result = await client.PUT("/approval-config/processes/{id}", {
      params: { path: { id } },
      body,
    });
    return unwrapOpenApiResponse(result);
  },

  async activateProcess(id: string): Promise<ApprovalProcessConfig> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/processes/{id}/activate", {
      params: { path: { id } },
    });
    return unwrapOpenApiResponse(result);
  },

  async deactivateProcess(id: string): Promise<ApprovalProcessConfig> {
    const client = useOpenApiClient();
    const result = await client.POST("/approval-config/processes/{id}/deactivate", {
      params: { path: { id } },
    });
    return unwrapOpenApiResponse(result);
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
