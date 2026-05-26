<script setup lang="ts">
import MarketDataPage from "~/features/market-data/MarketDataPage.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth"],
});
</script>

<template>
  <MarketDataPage />
</template>
