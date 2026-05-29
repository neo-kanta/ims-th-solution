/**
 * useDashboardWorkflow
 * ------------------------------------------------------------------
 * Reactive controller for the dashboard's workflow panel. Calls the real
 * backend workflow module (/workflow/day-states/{contractId}) via the
 * generated OpenAPI client. The backend remains the source of truth for
 * `allowedActions` and `blockingReasons` — the UI only renders them.
 *
 * Response envelope note: the IMS backend wraps every successful response
 * in `{ data: ... }` (httputil.OK) but Swagger annotations describe the
 * inner shape directly, so openapi-typescript generates the inner shape as
 * the response type. We unwrap defensively (same approach as dashboardApi).
 */
import { computed, readonly, ref } from "vue";

import { OpenApiRequestError, useOpenApiClient } from "~/api/openapi";
import type { components } from "~/api/ims-api";

export type WorkflowStateResponse =
  components["schemas"]["WorkflowStateResponse"];
export type WorkflowHistoryResponse = components["schemas"]["HistoryResponse"];
export type WorkflowTransitionEntry = components["schemas"]["TransitionEntry"];
export type WorkflowExecuteRequest =
  components["schemas"]["ExecuteTransitionRequest"];
export type WorkflowExecuteResponse =
  components["schemas"]["TransitionResponse"];

function unwrapEnvelope<T>(payload: unknown): T | undefined {
  if (
    payload !== null
    && typeof payload === "object"
    && "data" in (payload as Record<string, unknown>)
  ) {
    return (payload as { data?: T }).data;
  }
  return payload as T | undefined;
}

function todayBangkokIso(): string {
  // Asia/Bangkok calendar date (YYYY-MM-DD) for the workflow business day.
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Bangkok",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(new Date());
  const lookup = Object.fromEntries(parts.map((p) => [p.type, p.value]));
  return `${lookup.year}-${lookup.month}-${lookup.day}`;
}

export function useDashboardWorkflow() {
  const client = useOpenApiClient();

  const contractId = ref<string | null>(null);
  const businessDate = ref<string>(todayBangkokIso());

  const state = ref<WorkflowStateResponse | null>(null);
  const history = ref<WorkflowTransitionEntry[]>([]);

  const loadingState = ref(false);
  const loadingHistory = ref(false);
  const executing = ref(false);
  const error = ref<string | null>(null);
  const lastResult = ref<WorkflowExecuteResponse | null>(null);

  const hasContract = computed(() => Boolean(contractId.value));

  function setContract(id: string | null) {
    if (id === contractId.value) return;
    contractId.value = id;
    state.value = null;
    history.value = [];
    error.value = null;
    lastResult.value = null;
  }

  function setBusinessDate(date: string) {
    if (date === businessDate.value) return;
    businessDate.value = date;
    state.value = null;
    history.value = [];
    error.value = null;
  }

  function describeError(err: unknown, fallback: string): string {
    if (err instanceof OpenApiRequestError) return err.message;
    if (err instanceof Error) return err.message;
    return fallback;
  }

  async function fetchState() {
    const id = contractId.value;
    if (!id) return;
    loadingState.value = true;
    error.value = null;
    try {
      const response = await client.GET("/workflow/day-states/{contractId}", {
        params: {
          path: { contractId: id },
          query: { businessDate: businessDate.value },
        },
      });
      if (response.error !== undefined) {
        throw new OpenApiRequestError(response.response, response.error);
      }
      state.value =
        unwrapEnvelope<WorkflowStateResponse>(response.data) ?? null;
    } catch (err) {
      state.value = null;
      error.value = describeError(err, "Failed to load workflow state");
    } finally {
      loadingState.value = false;
    }
  }

  async function fetchHistory() {
    const id = contractId.value;
    if (!id) return;
    loadingHistory.value = true;
    try {
      const response = await client.GET(
        "/workflow/day-states/{contractId}/history",
        {
          params: {
            path: { contractId: id },
            query: { businessDate: businessDate.value },
          },
        },
      );
      if (response.error !== undefined) {
        throw new OpenApiRequestError(response.response, response.error);
      }
      const body =
        unwrapEnvelope<WorkflowHistoryResponse>(response.data) ?? {};
      history.value = body.transitions ?? [];
    } catch {
      // History is auxiliary; don't surface its failure as the main error.
      history.value = [];
    } finally {
      loadingHistory.value = false;
    }
  }

  async function refresh() {
    await Promise.all([fetchState(), fetchHistory()]);
  }

  async function execute(payload: WorkflowExecuteRequest): Promise<boolean> {
    const id = contractId.value;
    if (!id || executing.value) return false;
    executing.value = true;
    error.value = null;
    try {
      const response = await client.POST(
        "/workflow/day-states/{contractId}/transitions",
        {
          params: { path: { contractId: id } },
          body: payload,
        },
      );
      if (response.error !== undefined) {
        throw new OpenApiRequestError(response.response, response.error);
      }
      lastResult.value =
        unwrapEnvelope<WorkflowExecuteResponse>(response.data) ?? null;
      await refresh();
      return true;
    } catch (err) {
      error.value = describeError(err, "Workflow operation failed");
      return false;
    } finally {
      executing.value = false;
    }
  }

  function clearLastResult() {
    lastResult.value = null;
  }

  return {
    contractId: readonly(contractId),
    businessDate: readonly(businessDate),
    state: readonly(state),
    history: readonly(history),
    loadingState: readonly(loadingState),
    loadingHistory: readonly(loadingHistory),
    executing: readonly(executing),
    error: readonly(error),
    lastResult: readonly(lastResult),
    hasContract,
    setContract,
    setBusinessDate,
    fetchState,
    fetchHistory,
    refresh,
    execute,
    clearLastResult,
  };
}
