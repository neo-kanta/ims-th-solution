<script setup lang="ts">
import PortfolioConfirmationsView from "~/features/portfolio-workspace/PortfolioConfirmationsView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_CONFIRMATION_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioConfirmationsView :portfolio-code="portfolioCode" />
</template>
