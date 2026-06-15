import { ref, watch } from "vue";
import { workflowApi } from "../api/workflow.api";
import type { components } from "~/api/ims-api";
import { OpenApiRequestError } from "~/api/openapi";

export function todayBangkokIso(): string {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Bangkok",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(new Date());
  const lookup = Object.fromEntries(parts.map((p) => [p.type, p.value]));
  return `${lookup.year}-${lookup.month}-${lookup.day}`;
}

export function useWorkflowDaily() {
  const businessDate = ref<string>(todayBangkokIso());
  const dailyState = ref<components["schemas"]["DailyWorkflowResponse"] | null>(null);
  const loading = ref(false);
  const executing = ref(false);
  const error = ref<string | null>(null);

  async function fetchState() {
    loading.value = true;
    error.value = null;
    try {
      const state = await workflowApi.getDailyWorkflow(businessDate.value);
      dailyState.value = state;
    } catch (err: any) {
      dailyState.value = null;
      error.value = err instanceof OpenApiRequestError ? err.message : (err?.message || "Failed to load daily workflow state");
    } finally {
      loading.value = false;
    }
  }

  async function execute(
    operationType: string,
    remark?: string,
    zeroTransactionAttestation?: boolean,
    attestationReason?: string,
    notes?: string,
  ): Promise<boolean> {
    if (executing.value) return false;
    executing.value = true;
    error.value = null;
    try {
      const payload: components["schemas"]["DailyExecuteRequest"] = {
        businessDate: businessDate.value,
        operationType,
      };
      if (remark && remark.trim()) {
        payload.remark = remark.trim();
      }
      if (zeroTransactionAttestation !== undefined) {
        payload.zeroTransactionAttestation = zeroTransactionAttestation;
      }
      if (attestationReason && attestationReason.trim()) {
        payload.attestationReason = attestationReason.trim();
      }
      if (notes && notes.trim()) {
        payload.notes = notes.trim();
      }

      await workflowApi.executeDailyTransition(payload);
      await fetchState();
      return true;
    } catch (err: any) {
      error.value = err instanceof OpenApiRequestError ? err.message : (err?.message || "Failed to execute transition");
      return false;
    } finally {
      executing.value = false;
    }
  }

  watch(businessDate, () => {
    void fetchState();
  });

  return {
    businessDate,
    dailyState,
    loading,
    executing,
    error,
    fetchState,
    execute,
  };
}

