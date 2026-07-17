import { ref } from "vue";

import {
  portfolioDecisionApi,
  type ApiCreateDecisionV2Request,
  type ApiDecisionV2,
} from "../services/portfolioDecisionApi";
import { describeDecisionError } from "../lib/decisionErrors";

export type DecisionLifecyclePhase =
  | "idle"
  | "busy"
  | "draft-saved"
  | "submitted"
  | "create-failed"
  | "submit-failed";

/**
 * Orchestrates the two write flows the order ticket offers:
 *
 *   Save draft:            POST decisions                              -> DRAFT
 *   Submit for approval:   POST decisions, then POST decisions/{id}/submit
 *
 * "Submit Decision" never calls a ledger/transaction-posting endpoint — it
 * only starts the compliance/approval workflow (docs/api/portfolio-v2-api-ddd.md).
 *
 * Partial-failure handling: if create succeeds but submit fails, the
 * decision id is kept so retrySubmit() can retry the submit step alone —
 * create is never called again for the same attempt, so a form resubmission
 * cannot spawn a duplicate draft. `busy` gates every action so a duplicate
 * click cannot start a second in-flight request.
 */
export function useDecisionLifecycle() {
  const phase = ref<DecisionLifecyclePhase>("idle");
  const busy = ref(false);
  const error = ref<string | null>(null);
  const decisionId = ref<string | null>(null);
  const decision = ref<ApiDecisionV2 | null>(null);

  function reset() {
    phase.value = "idle";
    busy.value = false;
    error.value = null;
    decisionId.value = null;
    decision.value = null;
  }

  async function saveDraft(
    portfolioCode: string,
    body: ApiCreateDecisionV2Request,
  ): Promise<ApiDecisionV2 | null> {
    if (busy.value) return null;
    busy.value = true;
    phase.value = "busy";
    error.value = null;
    try {
      const created = await portfolioDecisionApi.create(portfolioCode, body);
      decisionId.value = created.id ?? null;
      decision.value = created;
      phase.value = "draft-saved";
      return created;
    } catch (err) {
      error.value = describeDecisionError(err, "Failed to save the draft decision.");
      phase.value = "create-failed";
      return null;
    } finally {
      busy.value = false;
    }
  }

  async function createThenSubmit(
    portfolioCode: string,
    body: ApiCreateDecisionV2Request,
  ): Promise<ApiDecisionV2 | null> {
    if (busy.value) return null;
    busy.value = true;
    phase.value = "busy";
    error.value = null;
    try {
      const created = await portfolioDecisionApi.create(portfolioCode, body);
      decisionId.value = created.id ?? null;
      decision.value = created;
    } catch (err) {
      error.value = describeDecisionError(err, "Failed to create the decision.");
      phase.value = "create-failed";
      busy.value = false;
      return null;
    }

    try {
      const submitted = await portfolioDecisionApi.submit(portfolioCode, decisionId.value!);
      decision.value = submitted;
      phase.value = "submitted";
      return submitted;
    } catch (err) {
      // The draft is already persisted — say so, and let the caller retry
      // the submit step alone via retrySubmit() using the stored decisionId.
      // describeDecisionError's fallback only fires when the backend gave no
      // message at all, so the "draft was saved" framing is prefixed
      // unconditionally rather than passed as that fallback.
      error.value = `The draft was saved, but submitting it for approval failed: ${describeDecisionError(err, "unknown error")}`;
      phase.value = "submit-failed";
      return null;
    } finally {
      busy.value = false;
    }
  }

  async function retrySubmit(portfolioCode: string): Promise<ApiDecisionV2 | null> {
    if (busy.value || !decisionId.value) return null;
    busy.value = true;
    error.value = null;
    try {
      const submitted = await portfolioDecisionApi.submit(portfolioCode, decisionId.value);
      decision.value = submitted;
      phase.value = "submitted";
      return submitted;
    } catch (err) {
      error.value = describeDecisionError(err, "Submit for approval failed again.");
      phase.value = "submit-failed";
      return null;
    } finally {
      busy.value = false;
    }
  }

  return {
    phase,
    busy,
    error,
    decisionId,
    decision,
    saveDraft,
    createThenSubmit,
    retrySubmit,
    reset,
  };
}
