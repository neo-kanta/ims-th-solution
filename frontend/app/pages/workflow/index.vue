<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "WORKFLOW_VIEW",
});

const { t } = useI18n();
</script>

<template>
  <div class="page">
    <AppPageHeader
      :title="t('navigation.workflow')"
      :description="t('workflow.subtitle')"
    />

    <div class="card">
      <div class="card-body">
        <p class="text-secondary">{{ t("dashboard.noData") }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  display: grid;
  gap: var(--space-7);
}
</style>
