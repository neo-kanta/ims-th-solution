<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.investmentReview.title"
    description-key="placeholders.investmentReview.description"
  />
</template>
