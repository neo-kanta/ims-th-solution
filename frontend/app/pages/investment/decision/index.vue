<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.investmentDecisions.title"
    description-key="placeholders.investmentDecisions.description"
  />
</template>
