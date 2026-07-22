<script setup lang="ts">
import PortfolioValuationsView from "~/features/portfolio-workspace/PortfolioValuationsView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_VALUATION_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioValuationsView :portfolio-code="portfolioCode" />
</template>
