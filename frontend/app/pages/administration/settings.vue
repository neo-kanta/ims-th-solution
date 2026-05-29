<script setup lang="ts">
import SettingsControlCenter from "../../features/settings/components/SettingsControlCenter.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth"],
});
</script>

<template>
  <SettingsControlCenter mode="administration" />
</template>
