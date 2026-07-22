<script setup lang="ts">
import PortfolioSettingsView from "~/features/portfolio-workspace/PortfolioSettingsView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  // Page-level gate is view-only, matching the backend's GET portfolio-detail
  // requirement (INVESTMENT_PORTFOLIO_VIEW) — the Save button itself is
  // separately gated on INVESTMENT_PORTFOLIO_MANAGE via IMSPermissionGuard
  // inside PortfolioSettingsView, matching confirmations.vue's convention.
  permission: "INVESTMENT_PORTFOLIO_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioSettingsView :portfolio-code="portfolioCode" />
</template>
