<script setup lang="ts">
import PortfolioDecisionNewView from "~/features/portfolio-decision/PortfolioDecisionNewView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_DECISION_MANAGE",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioDecisionNewView :portfolio-code="portfolioCode" />
</template>
