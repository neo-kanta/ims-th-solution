/**
 * Thin composable bridging the workflow Pinia store to consumers.
 *
 * Use this from pages/components rather than calling the store directly when
 * you want auto-binding of the OpenAPI client and a few convenience computed
 * props. The store remains the source of truth for state.
 */
import { computed } from "vue";

import { useOpenApiClient } from "~/api/openapi";

import { ACTION_CATALOG } from "../permissions";
import { useWorkflowStore } from "../store/useWorkflowStore";
import type { WorkflowAction } from "../types";

export function useWorkflowOperations() {
  const store = useWorkflowStore();
  if (!store.client) {
    store.setClient(useOpenApiClient());
  }

  const visibleActions = computed(() => {
    const allowed = new Set<string>(store.allowedActions);
    return (Object.values(ACTION_CATALOG)).filter((opt) =>
      allowed.has(opt.value),
    );
  });

  function actionOption(action: WorkflowAction) {
    return ACTION_CATALOG[action];
  }

  return {
    store,
    visibleActions,
    actionOption,
  };
}
