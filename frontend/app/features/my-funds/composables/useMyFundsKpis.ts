/**
 * Computes the top KPI strip from the decorated MyFundCard list.
 *
 * All numbers are derived from backend-provided authoritative values
 * (FundResponse.status, ValuationResponse.aum/unrealised_pnl, Breach.status,
 * WorkflowStateResponse.currentState). No invented data.
 */
import { computed, type Ref } from "vue";

import type { MyFundCard, MyFundsKpiStrip } from "../types";

const PENDING_WORKFLOW_STATES = new Set(["DAY_OPEN", "TRANSACTION_CLOSED"]);

export function useMyFundsKpis(cards: Ref<MyFundCard[]>) {
  const kpis = computed<MyFundsKpiStrip>(() => {
    let aum = 0;
    let unrealised = 0;
    let active = 0;
    let breachCount = 0;
    let staleCount = 0;
    let pending = 0;
    let worstSeverity: MyFundsKpiStrip["worst_breach_severity"] = null;
    const ccyTally: Record<string, number> = {};

    for (const card of cards.value) {
      aum += card.valuation.aum_numeric;
      unrealised += card.valuation.unrealised_pnl_numeric;
      breachCount += card.compliance.open_count;
      if (card.status === "ACTIVE" || card.status === "LOCKED" || card.status === "BREACH") {
        active += 1;
      }
      if (card.valuation.has_stale_inputs || !card.valuation.available) {
        staleCount += 1;
      }
      if (PENDING_WORKFLOW_STATES.has(card.workflow.current_state)) {
        pending += 1;
      }
      if (card.compliance.worst_severity) {
        if (
          card.compliance.worst_severity === "BLOCK"
          || (card.compliance.worst_severity === "WARN" && worstSeverity !== "BLOCK")
          || (card.compliance.worst_severity === "INFO" && worstSeverity === null)
        ) {
          worstSeverity = card.compliance.worst_severity;
        }
      }
      const ccy = card.valuation.valuation_ccy || card.base_currency || "";
      if (ccy) ccyTally[ccy] = (ccyTally[ccy] ?? 0) + 1;
    }

    const ccy = pickDominantCurrency(ccyTally) || cards.value[0]?.base_currency || "";
    const trend: MyFundsKpiStrip["unrealised_pnl_trend"] =
      unrealised > 0 ? "up" : unrealised < 0 ? "down" : "flat";

    return {
      total_aum: aum.toString(),
      total_aum_numeric: aum,
      total_unrealised_pnl: unrealised.toString(),
      total_unrealised_pnl_numeric: unrealised,
      unrealised_pnl_trend: trend,
      active_count: active,
      total_count: cards.value.length,
      open_breach_count: breachCount,
      worst_breach_severity: worstSeverity,
      stale_count: staleCount,
      pending_workflow_count: pending,
      valuation_ccy: ccy,
    };
  });

  return { kpis };
}

function pickDominantCurrency(tally: Record<string, number>): string {
  let best = "";
  let bestCount = -1;
  for (const [ccy, count] of Object.entries(tally)) {
    if (count > bestCount) {
      best = ccy;
      bestCount = count;
    }
  }
  return best;
}
