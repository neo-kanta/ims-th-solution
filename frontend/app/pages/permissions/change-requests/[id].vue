<script setup lang="ts">
import PermissionRequestDetailScreen from "~/features/permissions/components/PermissionRequestDetailScreen.vue";

const route = useRoute();

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "permission.change_request.review",
});
</script>

<template>
  <PermissionRequestDetailScreen :request-id="String(route.params.id)" />
</template>
