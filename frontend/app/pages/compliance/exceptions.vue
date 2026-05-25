<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import ComplianceExceptionForm from "~/features/compliance/components/ComplianceExceptionForm.vue";
import ComplianceExceptionTimeline from "~/features/compliance/components/ComplianceExceptionTimeline.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});

const { t } = useI18n();
</script>

<template>
  <section class="exceptions-page">
    <AppPageHeader
      :title="t('compliance.exceptions.title')"
      :description="t('compliance.exceptions.description')"
    />

    <div class="exceptions-page__notice" role="status">
      {{ t("compliance.exceptions.notice") }}
    </div>

    <div class="exceptions-page__layout">
      <ComplianceExceptionForm />
      <ComplianceExceptionTimeline />
    </div>
  </section>
</template>

<style scoped>
.exceptions-page {
  display: grid;
  gap: var(--space-7);
}

.exceptions-page__notice {
  padding: var(--space-4) var(--space-5);
  border: 1px solid var(--alert-warning-border);
  border-radius: var(--radius-md);
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  font-size: var(--font-size-sm);
}

.exceptions-page__layout {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(220px, 1fr);
  gap: var(--space-6);
  align-items: start;
}

@media (max-width: 1024px) {
  .exceptions-page__layout {
    grid-template-columns: 1fr;
  }
}
</style>
