<script setup lang="ts">
// Redirect bare portfolio route → overview.
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_PORTFOLIO_VIEW",
});

const route = useRoute();
const router = useRouter();

const portfolioCode = String(route.params.portfolioCode ?? "");

onMounted(() => {
  if (portfolioCode) {
    void router.replace(`/portfolios/${portfolioCode}/overview`);
  }
});
</script>

<template>
  <div />
</template>
