<script setup lang="ts">
import SettingsControlCenter from "../../features/settings/components/SettingsControlCenter.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IAM_USER_VIEW",
});
</script>

<template>
  <SettingsControlCenter mode="administration" />
</template>
