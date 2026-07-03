/**
 * Thin composable that bridges the workflow Pinia store to the dashboard
 * workflow panel. Wraps useWorkflowStore so the panel never imports the store
 * directly and we can test the panel in isolation.
 */
import { storeToRefs } from "pinia";
import { useOpenApiClient } from "~/api/openapi";
import {
  useWorkflowStore,
  type WorkflowExecuteParams,
} from "~/features/workflow/store/useWorkflowStore";
export type {
  WorkflowExecuteRequest,
  WorkflowStateResponse,
} from "~/features/workflow/types";

import type { WorkflowAction, WorkflowExecuteRequest } from "~/features/workflow/types";

export function useDashboardWorkflow() {
  const store = useWorkflowStore();
  if (!store.client) {
    store.setClient(useOpenApiClient());
  }

  const {
    businessDate,
    state,
    history,
    loadingState,
    loadingHistory,
    executing,
    error,
    lastResult,
  } = storeToRefs(store);

  function setBusinessDate(date: string) {
    store.setBusinessDate(date);
  }

  function clearLastResult() {
    store.clearLastResult();
  }

  async function refresh() {
    await store.refresh();
  }

  async function execute(payload: {
    operationType: WorkflowAction;
    reason?: string;
    notes?: string;
    zeroTransactionAttestation?: boolean;
    attestationReason?: string;
  }): Promise<boolean> {
    const params: WorkflowExecuteParams = {
      action: payload.operationType,
    };
    if (payload.reason) params.reason = payload.reason;
    if (payload.notes) params.notes = payload.notes;
    if (payload.zeroTransactionAttestation !== undefined) {
      params.zeroTransactionAttestation = payload.zeroTransactionAttestation;
    }
    if (payload.attestationReason) {
      params.attestationReason = payload.attestationReason;
    }
    return store.execute(params);
  }

  return {
    businessDate,
    state,
    history,
    loadingState,
    loadingHistory,
    executing,
    error,
    lastResult,
    setBusinessDate,
    refresh,
    execute,
    clearLastResult,
  };
}
