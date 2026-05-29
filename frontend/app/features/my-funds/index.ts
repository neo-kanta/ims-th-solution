export { useMyFunds } from "./composables/useMyFunds";
export { useMyFundsKpis } from "./composables/useMyFundsKpis";
export * from "./types";
export {
  activeStageIndex,
  aggregateValuations,
  deriveRole,
  deriveStatus,
  emptyValuationSummary,
  mapWorkflowSummary,
  summarizeBreaches,
  todayBangkokIso,
  WORKFLOW_STAGES,
} from "./lib/derive";
export type { WorkflowStageKey } from "./lib/derive";
export {
  currencySymbol,
  formatMoneyCompact,
  formatMoneyExact,
  formatPercent,
  parseDecimal,
  parseDecimalOrNull,
  relativeTime,
} from "./lib/format";
export { myFundsApi } from "./services/myFundsApi";

export { default as FundSummaryCard } from "./components/FundSummaryCard.vue";
export { default as FundKpiStrip } from "./components/FundKpiStrip.vue";
export { default as FundWorkflowBar } from "./components/FundWorkflowBar.vue";
export { default as ComplianceBadge } from "./components/ComplianceBadge.vue";
export { default as RoleActionMenu } from "./components/RoleActionMenu.vue";
export { default as FundFilterToolbar } from "./components/FundFilterToolbar.vue";

export { default as FundDetailLayout } from "./components/detail/FundDetailLayout.vue";
export { default as FundDetailHeader } from "./components/detail/FundDetailHeader.vue";
export { default as StagesPanel } from "./components/detail/StagesPanel.vue";
export { default as CompliancePanel } from "./components/detail/CompliancePanel.vue";
export { default as AuditPanel } from "./components/detail/AuditPanel.vue";
export { default as EmptyTabPanel } from "./components/detail/EmptyTabPanel.vue";

export { useFundDetail } from "./composables/useFundDetail";
