<script setup lang="ts">
import DashboardOverviewScreen from "../features/dashboard/components/DashboardOverviewScreen.vue";

definePageMeta({
  layout: "dashboard",
  middleware: "auth",
});
</script>

<template>
  <DashboardOverviewScreen />
</template>
