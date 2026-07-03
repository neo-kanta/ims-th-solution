<script setup lang="ts">
import PersonalSettingsConsole from "../../features/settings/components/PersonalSettingsConsole.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth"],
});
</script>

<template>
  <PersonalSettingsConsole />
</template>
