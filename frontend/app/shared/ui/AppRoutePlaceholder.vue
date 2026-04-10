<script setup lang="ts">
import type { AppTranslationKey } from "../i18n/messages";

interface Props {
  titleKey: AppTranslationKey;
  descriptionKey: AppTranslationKey;
}

const props = defineProps<Props>();
const { t } = useI18n();
</script>

<template>
  <div class="route-placeholder">
    <div class="route-placeholder__header">
      <h1 class="route-placeholder__title">{{ t(props.titleKey) }}</h1>
      <p class="route-placeholder__copy">{{ t(props.descriptionKey) }}</p>
    </div>

    <AppCard class="route-placeholder__card">
      <div class="route-placeholder__body">
        <AppIcon name="info" size="lg" />
        <p class="route-placeholder__copy">{{ t(props.descriptionKey) }}</p>
      </div>
    </AppCard>
  </div>
</template>

<style scoped>
.route-placeholder {
  display: grid;
  gap: var(--space-6);
}

.route-placeholder__header {
  display: grid;
  gap: var(--space-2);
}

.route-placeholder__title {
  margin: 0;
  color: var(--text-primary);
  font-size: clamp(1.75rem, 2vw, 2.125rem);
  font-weight: var(--font-weight-bold);
  letter-spacing: -0.03em;
}

.route-placeholder__copy {
  margin: 0;
  color: var(--text-secondary);
  max-width: 36rem;
}

.route-placeholder__card {
  min-height: 12rem;
}

.route-placeholder__body {
  min-height: 12rem;
  display: grid;
  gap: var(--space-4);
  place-items: center;
  text-align: center;
  padding: var(--space-8);
}
</style>
