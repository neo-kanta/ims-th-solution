import { ref } from "vue";
import { workflowApi } from "../api/workflow.api";
import type { components } from "~/api/ims-api";
import { OpenApiRequestError } from "~/api/openapi";

export function useWorkflowSettings() {
  const settings = ref<components["schemas"]["OperationSettingEntry"][]>([]);
  const loading = ref(false);
  const saving = ref(false);
  const error = ref<string | null>(null);

  async function fetchSettings() {
    loading.value = true;
    error.value = null;
    try {
      const response = await workflowApi.getWorkflowSettings();
      settings.value = response.settings ?? [];
    } catch (err: any) {
      settings.value = [];
      error.value = err instanceof OpenApiRequestError ? err.message : (err?.message || "Failed to load settings");
    } finally {
      loading.value = false;
    }
  }

  async function saveSettings(
    operationType: string,
    approvers: components["schemas"]["ApproverInput"][],
  ): Promise<boolean> {
    saving.value = true;
    error.value = null;
    try {
      const payload: components["schemas"]["DailySettingsUpdateRequest"] = {
        operationType,
        approvers,
      };
      await workflowApi.updateWorkflowSettings(payload);
      await fetchSettings();
      return true;
    } catch (err: any) {
      error.value = err instanceof OpenApiRequestError ? err.message : (err?.message || "Failed to save settings");
      return false;
    } finally {
      saving.value = false;
    }
  }

  return {
    settings,
    loading,
    saving,
    error,
    fetchSettings,
    saveSettings,
  };
}
