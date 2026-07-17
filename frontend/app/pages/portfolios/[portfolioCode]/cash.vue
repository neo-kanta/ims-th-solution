<script setup lang="ts">
import PortfolioCashView from "~/features/portfolio-workspace/PortfolioCashView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_PORTFOLIO_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioCashView :portfolio-code="portfolioCode" />
</template>
