/**
 * Workflow Pinia store.
 *
 * Holds the daily workflow state and active businessDate, the latest
 * `WorkflowStateResponse`, the transition history, loading flags, and the
 * last error / result. Actions wrap the typed OpenAPI client so components
 * never speak to fetch directly.
 *
 * The store dependency-injects its API client via `setClient(...)` so unit
 * tests can swap a fake without needing the Nuxt runtime. In the live app
 * the layout/page calls `ensureClient()` once on mount to bind the real
 * `useOpenApiClient()` instance.
 *
 * Backend remains the source of truth for transitions and `allowedActions`.
 * The store does not re-derive what is/isn't permitted; it surfaces what
 * the backend reports and forwards any execute requests verbatim.
 */
import { defineStore } from "pinia";

import { OpenApiRequestError, unwrapEnvelope } from "~/api/openapi";
import type { ImsOpenApiClient } from "~/api/openapi";

import type {
  WorkflowAction,
  WorkflowExecuteRequest,
  WorkflowExecuteResponse,
  WorkflowHistoryResponse,
  WorkflowStateResponse,
  WorkflowTransitionEntry,
} from "../types";

export function todayBangkokIso(): string {
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

function newIdempotencyKey(): string {
  if (typeof globalThis.crypto?.randomUUID === "function") {
    return globalThis.crypto.randomUUID();
  }
  // Deterministic-enough fallback for non-crypto environments (very old node
  // in unit tests). Never reached in modern browsers / Node 19+.
  return `wf-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

export interface WorkflowExecuteParams {
  action: WorkflowAction;
  reason?: string;
  notes?: string;
  zeroTransactionAttestation?: boolean;
  attestationReason?: string;
}

const clients = new WeakMap<any, ImsOpenApiClient | null>();

interface WorkflowStoreState {
  businessDate: string;
  state: WorkflowStateResponse | null;
  history: WorkflowTransitionEntry[];
  loadingState: boolean;
  loadingHistory: boolean;
  executing: boolean;
  error: string | null;
  lastResult: WorkflowExecuteResponse | null;
  lastIdempotencyKey: string | null;
}

export const useWorkflowStore = defineStore("workflow", {
  state: (): WorkflowStoreState => ({
    businessDate: todayBangkokIso(),
    state: null,
    history: [],
    loadingState: false,
    loadingHistory: false,
    executing: false,
    error: null,
    lastResult: null,
    lastIdempotencyKey: null,
  }),

  getters: {
    client(): ImsOpenApiClient | null {
      return clients.get(this) ?? null;
    },
    isPersisted: (s) => Boolean(s.state?.persisted),
    currentStateCode: (s) => s.state?.currentState ?? "NOT_STARTED",
    allowedActions: (s): readonly string[] => s.state?.allowedOperations ?? [],
    blockingReasons: (s) => s.state?.blockedReasons ?? [],
    transactionsLocked: (s) =>
      s.state?.currentState === "MANAGER_APPROVED"
      || s.state?.currentState === "TRANSACTION_CLOSED"
      || s.state?.currentState === "ACCOUNTING_CLOSED",
    /** True after manager approval — the investment workspace should lock. */
    managerApproved: (s) =>
      s.state?.currentState === "MANAGER_APPROVED"
      || s.state?.currentState === "TRANSACTION_CLOSED"
      || s.state?.currentState === "ACCOUNTING_CLOSED",
  },

  actions: {
    /** Inject the typed OpenAPI client. Called from app code; tests pass a fake. */
    setClient(client: ImsOpenApiClient | null) {
      clients.set(this, client);
    },

    setBusinessDate(date: string) {
      if (date === this.businessDate) return;
      this.businessDate = date;
      this.state = null;
      this.history = [];
      this.error = null;
    },

    clearLastResult() {
      this.lastResult = null;
    },

    clearError() {
      this.error = null;
    },

    requireClient(): ImsOpenApiClient {
      if (!this.client) {
        throw new Error(
          "workflow store: OpenAPI client not bound. Call setClient(useOpenApiClient()) first.",
        );
      }
      return this.client;
    },

    describeError(err: unknown, fallback: string): string {
      if (err instanceof OpenApiRequestError) return err.message;
      if (err instanceof Error) return err.message;
      return fallback;
    },

    async fetchState() {
      this.loadingState = true;
      this.error = null;
      try {
        const client = this.requireClient();
        const response = await client.GET(
          "/workflow/daily",
          {
            params: {
              query: {
                businessDate: this.businessDate,
              },
            },
          },
        );
        if (response.error !== undefined) {
          throw new OpenApiRequestError(response.response, response.error);
        }
        this.state =
          unwrapEnvelope<WorkflowStateResponse>(response.data) ?? null;
      } catch (err) {
        this.state = null;
        this.error = this.describeError(err, "Failed to load workflow state");
      } finally {
        this.loadingState = false;
      }
    },

    async fetchHistory() {
      this.loadingHistory = true;
      try {
        const client = this.requireClient();
        const response = await client.GET(
          "/workflow/daily/transitions",
          {
            params: {
              query: {
                businessDate: this.businessDate,
                page: 1,
                pageSize: 50,
              },
            },
          },
        );
        if (response.error !== undefined) {
          throw new OpenApiRequestError(response.response, response.error);
        }
        const body =
          unwrapEnvelope<WorkflowHistoryResponse>(response.data) ?? {};
        this.history = body.transitions ?? [];
      } catch {
        // History is auxiliary; don't surface its failure as the main error.
        this.history = [];
      } finally {
        this.loadingHistory = false;
      }
    },

    async refresh() {
      await Promise.all([this.fetchState(), this.fetchHistory()]);
    },

    /**
     * Execute a workflow transition. Returns true on success, false on
     * error (the error is captured in `state.error`). Double-submit is
     * guarded by `executing` — the second call resolves to false without
     * hitting the backend.
     */
    async execute(params: WorkflowExecuteParams): Promise<boolean> {
      if (this.executing) return false;
      this.executing = true;
      this.error = null;
      try {
        const body: WorkflowExecuteRequest = {
          businessDate: this.businessDate,
          operationType: params.action,
        };
        if (params.reason && params.reason.trim()) {
          body.remark = params.reason.trim();
        }
        if (params.notes && params.notes.trim()) {
          body.notes = params.notes.trim();
        }
        if (params.action === "MANAGER_APPROVE") {
          body.zeroTransactionAttestation = Boolean(
            params.zeroTransactionAttestation,
          );
          if (
            params.zeroTransactionAttestation
            && params.attestationReason
            && params.attestationReason.trim()
          ) {
            body.attestationReason = params.attestationReason.trim();
          }
        }

        const client = this.requireClient();
        const response = await client.POST(
          "/workflow/daily/execute",
          {
            body,
          },
        );
        if (response.error !== undefined) {
          throw new OpenApiRequestError(response.response, response.error);
        }
        this.lastResult =
          unwrapEnvelope<WorkflowExecuteResponse>(response.data) ?? null;
        await this.refresh();
        return true;
      } catch (err) {
        this.error = this.describeError(err, "Workflow operation failed");
        return false;
      } finally {
        this.executing = false;
      }
    },
  },
});

export type WorkflowStore = ReturnType<typeof useWorkflowStore>;
