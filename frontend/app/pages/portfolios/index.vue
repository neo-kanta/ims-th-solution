<script setup lang="ts">
import PortfolioDirectoryView from "~/features/portfolio-workspace/PortfolioDirectoryView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_PORTFOLIO_VIEW",
});
</script>

<template>
  <PortfolioDirectoryView />
</template>
