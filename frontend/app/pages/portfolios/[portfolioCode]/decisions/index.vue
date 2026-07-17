<script setup lang="ts">
import PortfolioDecisionsListView from "~/features/portfolio-decision/PortfolioDecisionsListView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_DECISION_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioDecisionsListView :portfolio-code="portfolioCode" />
</template>
