<script setup lang="ts">
import PortfolioWatchlistsView from "~/features/portfolio-workspace/PortfolioWatchlistsView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "WATCHLIST_VIEW",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioWatchlistsView :portfolio-code="portfolioCode" />
</template>
