<script setup lang="ts">
import PortfolioLedgerView from "~/features/portfolio-workspace/PortfolioLedgerView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_PORTFOLIO_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioLedgerView :portfolio-code="portfolioCode" />
</template>
