<script setup lang="ts">
/**
 * OP-01 "BUY/SELL Single Securities" entry point. Portfolio-first: it
 * queries the full portfolio list itself and lets the user pick one, rather
 * than requiring a fund to be pre-selected on the Operator catalog page.
 */
import PortfolioDecisionPickerView from "~/features/portfolio-decision/PortfolioDecisionPickerView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_DECISION_MANAGE",
});
</script>

<template>
  <PortfolioDecisionPickerView />
</template>
