<script setup lang="ts">
import MarketDataSecurityDetailPage from "~/features/market-data/MarketDataSecurityDetailPage.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth"],
});

const route = useRoute();
const securityId = computed(() => String(route.params.securityId ?? ""));
</script>

<template>
  <MarketDataSecurityDetailPage :security-id="securityId" />
</template>
