<script setup lang="ts">
// Redirect bare fund route → holdings.
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const route = useRoute();
const router = useRouter();

const fundId = String(route.params.fundId ?? "");

onMounted(() => {
  if (fundId) {
    void router.replace(`/investment/funds/${fundId}/holdings`);
  }
});
</script>

<template>
  <div />
</template>
