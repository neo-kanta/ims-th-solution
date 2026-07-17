<script setup lang="ts">
import PortfolioDecisionDetailView from "~/features/portfolio-decision/PortfolioDecisionDetailView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_DECISION_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
const decisionId = computed(() => String(route.params.decisionId ?? ""));
</script>

<template>
  <PortfolioDecisionDetailView :portfolio-code="portfolioCode" :decision-id="decisionId" />
</template>
