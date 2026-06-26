<script setup lang="ts">
import { useRoute, navigateTo } from "#imports";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  meta: { permission: "WATCHLIST_VIEW" },
});

const route = useRoute();
const alertId = route.params.id as string;
navigateTo(`/watchlist?alert_id=${alertId}`, { replace: true });
</script>

<template>
  <div />
</template>
