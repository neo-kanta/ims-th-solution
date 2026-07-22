import { ref } from "vue";

import { useComplianceCheckGroup } from "../../compliance/composables/useComplianceCheckGroup";
import { approvalApi } from "../../approval/services/approvalApi";
import type { ApprovalEvent } from "../../approval/types";
import { portfolioApi } from "../../portfolio-workspace/services/portfolioApi";
import type {
  ApiExecutionV2,
  ApiTradeConfirmationV2,
} from "../../portfolio-workspace/services/portfolioApi";
import type { ApiDecisionV2 } from "../services/portfolioDecisionApi";

/**
 * Loads the print-summary sections that a `DecisionResponse` does not carry
 * inline: the compliance check-group result (via
 * `compliance_check_group_id`), the approval timeline (via
 * `approval_request_id`), and — now that a typed contract exists —
 * execution/fill and trade-confirmation status (via
 * `GET /portfolios/{portfolioCode}/executions` matched by
 * `ExecutionResponse.decision_id`, then
 * `GET /portfolios/{portfolioCode}/confirmations` matched by
 * `TradeConfirmationResponse.execution_id`). All three are best-effort — a
 * missing id, or no matched execution/confirmation row, means the decision
 * never went through that step yet, rendered as an explicit "not recorded"
 * state, never an error or a fabricated value.
 */
export function usePrintSummarySections() {
  const compliance = useComplianceCheckGroup();
  const approvalHistory = ref<ApprovalEvent[]>([]);
  const approvalHistoryLoading = ref(false);
  const approvalHistoryError = ref<string | null>(null);

  const executions = ref<ApiExecutionV2[]>([]);
  const executionsLoading = ref(false);
  const executionsError = ref<string | null>(null);
  const confirmationsByExecutionId = ref<Record<string, ApiTradeConfirmationV2>>({});

  let requestSeq = 0;

  async function loadExecutionSummary(portfolioCode: string, decisionId: string) {
    const current = requestSeq;
    executionsLoading.value = true;
    executionsError.value = null;
    try {
      const list = await portfolioApi.listExecutions(portfolioCode, { limit: 200 });
      if (current !== requestSeq) return;
      const matched = (list.items ?? []).filter((execution) => execution.decision_id === decisionId);
      executions.value = matched;

      const executionIds = matched
        .map((execution) => execution.id)
        .filter((id): id is string => Boolean(id));
      if (executionIds.length > 0) {
        try {
          const confirmations = await portfolioApi.listConfirmations(portfolioCode, { limit: 200 });
          if (current !== requestSeq) return;
          const byExecutionId: Record<string, ApiTradeConfirmationV2> = {};
          for (const confirmation of confirmations.items ?? []) {
            if (confirmation.execution_id && executionIds.includes(confirmation.execution_id)) {
              byExecutionId[confirmation.execution_id] = confirmation;
            }
          }
          confirmationsByExecutionId.value = byExecutionId;
        } catch {
          // Confirmations failing to load does not invalidate the execution
          // rows already resolved — the print sheet renders "not recorded"
          // for confirmation status only.
          if (current === requestSeq) confirmationsByExecutionId.value = {};
        }
      } else {
        confirmationsByExecutionId.value = {};
      }
    } catch (err) {
      if (current !== requestSeq) return;
      executions.value = [];
      confirmationsByExecutionId.value = {};
      executionsError.value =
        err instanceof Error ? err.message : "Failed to load execution/confirmation status.";
    } finally {
      if (current === requestSeq) executionsLoading.value = false;
    }
  }

  async function load(decision: ApiDecisionV2 | null, portfolioCode?: string) {
    const current = ++requestSeq;
    compliance.clear();
    approvalHistory.value = [];
    approvalHistoryError.value = null;
    executions.value = [];
    executionsError.value = null;
    confirmationsByExecutionId.value = {};

    if (!decision) return;

    if (decision.compliance_check_group_id) {
      void compliance.fetch(decision.compliance_check_group_id);
    }

    if (decision.approval_request_id) {
      approvalHistoryLoading.value = true;
      try {
        const events = await approvalApi.timeline(decision.approval_request_id);
        if (current !== requestSeq) return;
        approvalHistory.value = events ?? [];
      } catch (err) {
        if (current !== requestSeq) return;
        approvalHistoryError.value =
          err instanceof Error ? err.message : "Failed to load the approval history.";
      } finally {
        if (current === requestSeq) approvalHistoryLoading.value = false;
      }
    }

    if (portfolioCode && decision.id) {
      await loadExecutionSummary(portfolioCode, decision.id);
    }
  }

  return {
    compliance,
    approvalHistory,
    approvalHistoryLoading,
    approvalHistoryError,
    executions,
    executionsLoading,
    executionsError,
    confirmationsByExecutionId,
    load,
  };
}
