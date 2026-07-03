<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import CompliancePermissionMatrix from "~/features/compliance/components/CompliancePermissionMatrix.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});

const { t } = useI18n();
</script>

<template>
  <section class="perm-page">
    <AppPageHeader
      :title="t('compliance.permissions.title')"
      :description="t('compliance.permissions.description')"
    />
    <CompliancePermissionMatrix />
  </section>
</template>

<style scoped>
.perm-page {
  display: grid;
  gap: var(--space-7);
}
</style>
