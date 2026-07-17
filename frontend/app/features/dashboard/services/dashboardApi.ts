import {
  OpenApiRequestError,
  useOpenApiClient,
} from "~/api/openapi";
import type { components, paths } from "~/api/ims-api";

import type {
  ComplianceSeverity,
  DashboardSnapshotDTO,
  TaskDTO,
  TaskListDTO,
  TaskSummaryDTO,
  ValuationScope,
  ValuationSummaryDTO,
  WorkflowStateDTO,
} from "../types";
import { normalizeValuationSummary } from "../lib/valuationSummaryMapping";

// The IMS backend wraps every dashboard response in {data: …, message?: …}
// (see httputil.OK). openapi-typescript faithfully models that envelope on
// each route, and openapi-fetch surfaces the parsed envelope on
// `response.data`. We deliberately do NOT call `unwrapOpenApiResponse` for
// these endpoints because it expects the envelope's inner `data` field to be
// REQUIRED while Swagger marks it optional — the two type shapes don't line
// up. Unwrapping inline keeps the request fully type-safe and produces a
// strict, normalised value object for the caller.

type GeneratedTask = components["schemas"]["TaskDTO"];
type GeneratedWorkflowState = components["schemas"]["WorkflowStateDTO"];
type GeneratedSummary = components["schemas"]["TaskSummaryDTO"];
type GeneratedSnapshot = components["schemas"]["DashboardSnapshotDTO"];
type GeneratedTaskList = components["schemas"]["TaskListDTO"];

type SnapshotEnvelope =
  paths["/integration/dashboard/me"]["get"]["responses"][200]["content"]["application/json"];
type TaskListEnvelope =
  paths["/integration/tasks/my"]["get"]["responses"][200]["content"]["application/json"];
type SummaryEnvelope =
  paths["/integration/tasks/my/summary"]["get"]["responses"][200]["content"]["application/json"];
type ValuationSummaryEnvelope =
  paths["/integration/dashboard/valuation-summary"]["get"]["responses"][200]["content"]["application/json"];

const TASK_TYPES = ["RESEARCH_REVIEW", "WORKFLOW_PENDING", "COMPLIANCE_BREACH"] as const;
const TASK_PRIORITIES = ["HIGH", "MEDIUM", "LOW", "INFO"] as const;
const TASK_STATUSES = ["PENDING", "IN_PROGRESS", "COMPLETED"] as const;
const COMPLIANCE_SEVERITIES = ["BLOCK", "WARN", "REQUIRE_APPROVAL", "MONITOR"] as const;

function asSeverity(value: string | undefined): ComplianceSeverity {
  return (COMPLIANCE_SEVERITIES as readonly string[]).includes(value ?? "")
    ? (value as ComplianceSeverity)
    : "";
}

function asTaskType(value: string | undefined): TaskDTO["type"] {
  return (TASK_TYPES as readonly string[]).includes(value ?? "")
    ? (value as TaskDTO["type"])
    : "WORKFLOW_PENDING";
}

function asPriority(value: string | undefined): TaskDTO["priority"] {
  return (TASK_PRIORITIES as readonly string[]).includes(value ?? "")
    ? (value as TaskDTO["priority"])
    : "INFO";
}

function asStatus(value: string | undefined): TaskDTO["status"] {
  return (TASK_STATUSES as readonly string[]).includes(value ?? "")
    ? (value as TaskDTO["status"])
    : "PENDING";
}

function normalizeTask(task: GeneratedTask): TaskDTO {
  return {
    taskId: task.taskId ?? "",
    type: asTaskType(task.type),
    module: task.module ?? "",
    title: task.title ?? "",
    description: task.description ?? "",
    priority: asPriority(task.priority),
    status: asStatus(task.status),
    businessDate: task.businessDate ?? "",
    sourceRecordId: task.sourceRecordId ?? "",
    sourceType: task.sourceType ?? "",
    actionUrl: task.actionUrl ?? "",
    reason: task.reason ?? "",
    subject: task.subject ?? "",
    severity: asSeverity(task.severity),
    canAct: Boolean(task.canAct),
    allowedActions: task.allowedActions ?? [],
    contractId: task.contractId ?? "",
    createdAt: task.createdAt ?? "",
    updatedAt: task.updatedAt ?? "",
  };
}

function normalizeWorkflowState(
  state: GeneratedWorkflowState,
): WorkflowStateDTO {
  return {
    contractId: state.contractId ?? "",
    businessDate: state.businessDate ?? "",
    currentState: state.currentState ?? "",
    updatedAt: state.updatedAt ?? "",
  };
}

function normalizeSummary(
  summary: GeneratedSummary | undefined,
): TaskSummaryDTO {
  return {
    total: summary?.total ?? 0,
    byModule: summary?.byModule ?? {},
    byPriority: summary?.byPriority ?? {},
    highPriority: summary?.highPriority ?? 0,
  };
}

function normalizeSnapshot(
  payload: GeneratedSnapshot | undefined,
): DashboardSnapshotDTO {
  return {
    tasks: (payload?.tasks ?? []).map(normalizeTask),
    summary: normalizeSummary(payload?.summary),
    workflowStates: (payload?.workflowStates ?? []).map(normalizeWorkflowState),
    lastRefreshed: payload?.lastRefreshed ?? "",
  };
}

function normalizeTaskList(payload: GeneratedTaskList | undefined): TaskListDTO {
  return {
    tasks: (payload?.tasks ?? []).map(normalizeTask),
    summary: normalizeSummary(payload?.summary),
  };
}

// Generic envelope-aware unwrap. Throws OpenApiRequestError when the request
// itself failed; tolerates a missing `data` field by passing `undefined` to
// the caller's normaliser (the normaliser is responsible for defaults).
function unwrapEnvelope<TEnvelope extends { data?: unknown }>(result: {
  data?: TEnvelope;
  error?: unknown;
  response: Response;
}): TEnvelope["data"] | undefined {
  if (result.error !== undefined) {
    throw new OpenApiRequestError(result.response, result.error);
  }
  return result.data?.data;
}

export interface MyTasksQuery {
  module?: string;
  priority?: TaskDTO["priority"];
  status?: TaskDTO["status"];
}

export const dashboardApi = {
  async snapshot(): Promise<DashboardSnapshotDTO> {
    const client = useOpenApiClient();
    const response = await client.GET("/integration/dashboard/me");

    return normalizeSnapshot(unwrapEnvelope<SnapshotEnvelope>(response));
  },

  async myTasks(query: MyTasksQuery = {}): Promise<TaskListDTO> {
    const client = useOpenApiClient();
    const response = await client.GET("/integration/tasks/my", {
      params: {
        query: {
          module: query.module,
          priority: query.priority,
          status: query.status,
        },
      },
    });

    return normalizeTaskList(unwrapEnvelope<TaskListEnvelope>(response));
  },

  async mySummary(): Promise<TaskSummaryDTO> {
    const client = useOpenApiClient();
    const response = await client.GET("/integration/tasks/my/summary");

    return normalizeSummary(unwrapEnvelope<SummaryEnvelope>(response));
  },

  async valuationSummary(scope: ValuationScope): Promise<ValuationSummaryDTO> {
    const client = useOpenApiClient();
    const response = await client.GET("/integration/dashboard/valuation-summary", {
      params: { query: { scope } },
    });

    return normalizeValuationSummary(unwrapEnvelope<ValuationSummaryEnvelope>(response));
  },
};
