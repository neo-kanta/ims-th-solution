<script setup lang="ts">
import PortfolioHoldingsView from "~/features/portfolio-workspace/PortfolioHoldingsView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_PORTFOLIO_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioHoldingsView :portfolio-code="portfolioCode" />
</template>
