<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppCard from "~/shared/ui/AppCard.vue";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";

const { t } = useI18n();
</script>

<template>
  <AppCard>
    <header class="rc-head">
      <h3 class="rc-title">
        {{ t("compliance.dashboard.recentChanges.title") }}
      </h3>
    </header>

    <AppEmptyState
      icon="folder"
      :title="t('compliance.dashboard.recentChanges.empty')"
    />
  </AppCard>
</template>

<style scoped>
.rc-head {
  margin-bottom: var(--space-3);
}

.rc-title {
  margin: 0;
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
}
</style>
