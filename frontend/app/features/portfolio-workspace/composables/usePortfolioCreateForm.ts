import { ref } from "vue";

import { portfolioApi, type ApiPortfolioV2 } from "../services/portfolioApi";
import {
  buildPortfolioCreateRequest,
  emptyPortfolioCreateFormValues,
  validatePortfolioCreateForm,
  type PortfolioCreateErrors,
  type PortfolioCreateFormValues,
} from "../lib/portfolioCreateValidation";
import {
  classifyPortfolioCreateError,
  type ClassifiedPortfolioCreateError,
} from "../lib/portfolioCreateErrors";

export type PortfolioCreatePhase = "idle" | "busy" | "success" | "failed";

/**
 * Orchestrates the Create Portfolio form: client-side validation,
 * duplicate-submission guard (`busy` mirrors `useDecisionLifecycle`'s
 * pattern — a second `submit()` call while one is in flight is a no-op),
 * and status-classified error surfacing. Entered values are never cleared
 * on a recoverable error (400/403/404/409/422) — only a successful create
 * or an explicit `reset()` clears the form.
 */
export function usePortfolioCreateForm() {
  const values = ref<PortfolioCreateFormValues>(emptyPortfolioCreateFormValues());
  const busy = ref(false);
  const phase = ref<PortfolioCreatePhase>("idle");
  const fieldErrors = ref<PortfolioCreateErrors>({});
  const submitError = ref<ClassifiedPortfolioCreateError | null>(null);
  const created = ref<ApiPortfolioV2 | null>(null);

  function setValues(next: Partial<PortfolioCreateFormValues>) {
    values.value = { ...values.value, ...next };
  }

  function reset() {
    values.value = emptyPortfolioCreateFormValues();
    busy.value = false;
    phase.value = "idle";
    fieldErrors.value = {};
    submitError.value = null;
    created.value = null;
  }

  async function submit(fallbackErrorMessage: string): Promise<ApiPortfolioV2 | null> {
    if (busy.value) return null;

    const validation = validatePortfolioCreateForm(values.value);
    fieldErrors.value = validation.errors;
    if (!validation.valid) {
      submitError.value = null;
      phase.value = "failed";
      return null;
    }

    busy.value = true;
    phase.value = "busy";
    submitError.value = null;
    try {
      const body = buildPortfolioCreateRequest(values.value);
      const portfolio = await portfolioApi.create(body);
      created.value = portfolio;
      phase.value = "success";
      return portfolio;
    } catch (err) {
      // Recoverable error — preserve `values` so the user does not have to
      // retype the form.
      submitError.value = classifyPortfolioCreateError(err, fallbackErrorMessage);
      phase.value = "failed";
      return null;
    } finally {
      busy.value = false;
    }
  }

  return {
    values,
    busy,
    phase,
    fieldErrors,
    submitError,
    created,
    setValues,
    submit,
    reset,
  };
}
