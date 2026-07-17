<script setup lang="ts">
import ComplianceControlCenterView from "~/features/compliance/ComplianceControlCenterView.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});
</script>

<template>
  <ComplianceControlCenterView />
</template>
