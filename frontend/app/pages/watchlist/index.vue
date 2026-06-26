<script setup lang="ts">
import WatchlistPage from "~/features/watchlist/WatchlistPage.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  meta: { permission: "WATCHLIST_VIEW" },
});
</script>

<template>
  <WatchlistPage />
</template>
