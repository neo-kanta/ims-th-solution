<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.executionOrders.title"
    description-key="placeholders.executionOrders.description"
  />
</template>
