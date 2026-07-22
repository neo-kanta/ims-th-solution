<script setup lang="ts">
import PortfolioCreateView from "~/features/portfolio-workspace/PortfolioCreateView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_PORTFOLIO_MANAGE",
});
</script>

<template>
  <PortfolioCreateView />
</template>
