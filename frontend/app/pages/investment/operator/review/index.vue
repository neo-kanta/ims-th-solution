<script setup lang="ts">
import PortfolioReviewPrintView from "~/features/portfolio-decision/PortfolioReviewPrintView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_DECISION_VIEW",
});
</script>

<template>
  <PortfolioReviewPrintView />
</template>
