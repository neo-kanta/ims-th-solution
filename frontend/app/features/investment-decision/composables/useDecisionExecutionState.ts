import { ref } from "vue";

import { portfolioApi } from "../../portfolio-workspace/services/portfolioApi";
import type { ApiExecutionV2 } from "../../portfolio-workspace/services/portfolioApi";

/**
 * Resolves the execution matched to each visible decision so OP-02 can show
 * a real execution-state column without mutating anything.
 *
 * `GET /portfolios/{portfolioCode}/executions` is scoped per portfolio, not
 * per decision, so this composable batches by the portfolio codes the
 * caller has already resolved (see `usePortfolioCodeLookup`) and matches
 * `ExecutionResponse.decision_id` back onto each decision client-side. A
 * portfolio's executions are only fetched once per module lifetime (cached
 * by code) — the workbench calls `loadForPortfolioCodes` again after every
 * list refresh, which is a cheap no-op for already-loaded codes.
 *
 * Simplification: if a decision has more than one execution row (e.g. a
 * retried/reissued execution), the most recently loaded one wins — this is
 * a read-only display aid, not the authoritative execution record.
 */
export function useDecisionExecutionState() {
  const executionsByDecisionId = ref<Record<string, ApiExecutionV2>>({});
  const loading = ref(false);
  const error = ref<string | null>(null);
  const loadedCodes = new Set<string>();

  async function loadForPortfolioCodes(codes: ReadonlyArray<string | undefined | null>): Promise<void> {
    const nonNullCodes = codes.filter((code): code is string => Boolean(code));
    const unique = Array.from(new Set(nonNullCodes)).filter((code) => !loadedCodes.has(code));
    if (unique.length === 0) return;

    loading.value = true;
    error.value = null;
    try {
      const results = await Promise.all(
        unique.map(async (code) => {
          try {
            const list = await portfolioApi.listExecutions(code, { limit: 200 });
            loadedCodes.add(code);
            return list.items ?? [];
          } catch {
            // A portfolio the caller cannot see (or a transient failure)
            // just leaves its decisions without a matched execution — the
            // grid renders those as decision-status-derived states.
            loadedCodes.add(code);
            return [];
          }
        }),
      );

      const next = { ...executionsByDecisionId.value };
      for (const items of results) {
        for (const execution of items) {
          if (execution.decision_id) next[execution.decision_id] = execution;
        }
      }
      executionsByDecisionId.value = next;
    } finally {
      loading.value = false;
    }
  }

  function executionFor(decisionId: string | null | undefined): ApiExecutionV2 | null {
    if (!decisionId) return null;
    return executionsByDecisionId.value[decisionId] ?? null;
  }

  return { executionsByDecisionId, loading, error, loadForPortfolioCodes, executionFor };
}
