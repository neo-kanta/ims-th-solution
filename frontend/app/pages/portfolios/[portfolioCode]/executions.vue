<script setup lang="ts">
import PortfolioExecutionsView from "~/features/portfolio-workspace/PortfolioExecutionsView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_EXECUTION_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioExecutionsView :portfolio-code="portfolioCode" />
</template>
