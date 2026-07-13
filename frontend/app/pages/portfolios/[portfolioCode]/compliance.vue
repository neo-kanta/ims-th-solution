<script setup lang="ts">
import PortfolioComplianceView from "~/features/portfolio-workspace/PortfolioComplianceView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});

const route = useRoute();
const portfolioCode = computed(() => String(route.params.portfolioCode ?? ""));
</script>

<template>
  <PortfolioComplianceView :portfolio-code="portfolioCode" />
</template>
