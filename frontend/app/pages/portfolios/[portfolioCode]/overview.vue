<script setup lang="ts">
import PortfolioOverviewView from "~/features/portfolio-workspace/PortfolioOverviewView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_PORTFOLIO_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioOverviewView :portfolio-code="portfolioCode" />
</template>
