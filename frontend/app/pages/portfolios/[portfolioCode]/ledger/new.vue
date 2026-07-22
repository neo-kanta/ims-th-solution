<script setup lang="ts">
import PortfolioLedgerNewView from "~/features/portfolio-workspace/PortfolioLedgerNewView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_LEDGER_SIMULATE",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioLedgerNewView :portfolio-code="portfolioCode" />
</template>
