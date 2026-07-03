import { computed, ref } from "vue";

import { useI18n } from "~/composables/useI18n";

import { complianceApi } from "../services/complianceApi";
import type {
  ComplianceBreach,
  ComplianceCheckGroupResult,
  CompliancePreTradeRequest,
  CompliancePreTradeResponse,
} from "../types";

function statusOf(err: unknown): number | null {
  if (!err || typeof err !== "object") return null;
  const status = (err as { status?: unknown; statusCode?: unknown }).status;
  if (typeof status === "number") return status;
  const statusCode = (err as { statusCode?: unknown }).statusCode;
  if (typeof statusCode === "number") return statusCode;
  return null;
}

function backendErrorMessage(err: unknown): string | null {
  if (!err || typeof err !== "object") return null;
  const data = (err as { data?: { error?: unknown; message?: unknown } }).data;
  if (data) {
    if (typeof data.error === "string" && data.error.trim()) return data.error;
    if (typeof data.message === "string" && data.message.trim()) return data.message;
  }
  const message = (err as { message?: unknown }).message;
  if (typeof message === "string" && message.trim()) return message;
  return null;
}

/**
 * Pre-trade compliance check composable — Phase 1.
 *
 * Holds the most recent request/response so the result panel can be
 * re-rendered without re-asking the user to fill the form, and the recheck
 * button can replay the exact payload.
 *
 * On non-PASS verdicts we follow up with `GET /compliance/checks/{groupID}`
 * to hydrate each breach with its persisted evidence object. The live
 * pre-trade response's `BreachSummary` lacks evidence; the persisted record
 * (which is the same data the post-trade inbox uses) carries
 * `threshold_breached.{actual, limit, operator, unit}` plus rule-specific
 * metrics. Without this step the BLOCK card cannot show real numbers, and
 * we are not permitted to invent them.
 */
export function useComplianceChecks() {
  const { t } = useI18n();
  const lastRequest = ref<CompliancePreTradeRequest | null>(null);
  const result = ref<CompliancePreTradeResponse | null>(null);
  const hydratedBreaches = ref<ComplianceBreach[]>([]);
  const hydrationError = ref<string | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  /** Map from breach_id → persisted breach (with evidence). */
  const breachById = computed(() => {
    const m = new Map<string, ComplianceBreach>();
    for (const b of hydratedBreaches.value) m.set(b.id, b);
    return m;
  });

  function classifyError(err: unknown): string {
    const status = statusOf(err);
    const backend = backendErrorMessage(err);

    if (status === 401) {
      return t("compliance.preTrade.errors.unauthorizedRules");
    }
    if (status === 403) {
      // The backend gates pre-trade checks on WORKFLOW_EXECUTE.
      // Spec demands a specific copy here, not the raw backend message.
      return t("compliance.preTrade.errors.workflowExecuteRequired");
    }
    const base = backend ?? "Pre-trade compliance check failed.";
    return status ? `${status} · ${base}` : base;
  }

  async function hydrate(groupId: string) {
    hydratedBreaches.value = [];
    hydrationError.value = null;
    try {
      const group = await complianceApi.getCheckGroup(groupId);
      hydratedBreaches.value = group.breaches ?? [];
    } catch (err) {
      // We do NOT fail the run on hydration error — the verdict is real,
      // we just can't show numbers for it. Surface the issue so the panel
      // can render the "Data not provided" hint with context.
      const status = statusOf(err);
      const backend = backendErrorMessage(err);
      hydrationError.value =
        (status ? `${status} · ` : "") +
        (backend ?? "Failed to fetch breach evidence.");
    }
  }

  async function runPreTrade(payload: CompliancePreTradeRequest) {
    loading.value = true;
    error.value = null;
    hydratedBreaches.value = [];
    hydrationError.value = null;
    try {
      lastRequest.value = payload;
      const res = await complianceApi.runPreTradeCheck(payload);
      result.value = res;

      // Hydrate the persisted breach evidence asynchronously when we have
      // a non-PASS verdict and at least one breach to look up. We await so
      // the result panel renders once with the real numbers, not twice.
      if (
        res.verdict !== "PASS" &&
        res.check_group_id &&
        Array.isArray(res.breaches) &&
        res.breaches.length > 0
      ) {
        await hydrate(res.check_group_id);
      }

      return res;
    } catch (err) {
      error.value = classifyError(err);
      result.value = null;
      throw err;
    } finally {
      loading.value = false;
    }
  }

  function reset() {
    result.value = null;
    error.value = null;
    hydratedBreaches.value = [];
    hydrationError.value = null;
  }

  return {
    lastRequest,
    result,
    hydratedBreaches,
    hydrationError,
    breachById,
    loading,
    error,
    runPreTrade,
    reset,
  };
}

/**
 * Check-group lookup — used by the Audit Trail page and the Post-trade inbox
 * detail drawer. The endpoint returns the persisted CheckRecord rows + the
 * Breach entities with structured evidence (which the live pre-trade response
 * does not include).
 */
export function useComplianceCheckGroup() {
  const result = ref<ComplianceCheckGroupResult | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchGroup(groupId: string) {
    loading.value = true;
    error.value = null;
    try {
      result.value = await complianceApi.getCheckGroup(groupId);
      return result.value;
    } catch (err) {
      const status = statusOf(err);
      const backend = backendErrorMessage(err);
      error.value =
        (status ? `${status} · ` : "") +
        (backend ?? "Check group lookup failed.");
      result.value = null;
      throw err;
    } finally {
      loading.value = false;
    }
  }

  function reset() {
    result.value = null;
    error.value = null;
  }

  return { result, loading, error, fetchGroup, reset };
}
