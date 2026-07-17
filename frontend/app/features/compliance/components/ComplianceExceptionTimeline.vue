<script setup lang="ts">
import { useI18n, type AppTranslationKey } from "~/composables/useI18n";
import AppCard from "~/shared/ui/AppCard.vue";

const { t } = useI18n();

interface Step {
  labelKey: AppTranslationKey;
  tone: "neutral" | "info" | "success" | "danger" | "warning";
}

const STEPS: readonly Step[] = [
  { labelKey: "compliance.exceptions.timeline.notRequested", tone: "neutral" },
  { labelKey: "compliance.exceptions.timeline.draft", tone: "neutral" },
  { labelKey: "compliance.exceptions.timeline.pending", tone: "info" },
  { labelKey: "compliance.exceptions.timeline.approved", tone: "success" },
  { labelKey: "compliance.exceptions.timeline.rejected", tone: "danger" },
  { labelKey: "compliance.exceptions.timeline.expired", tone: "warning" },
] as const;

function dotClass(tone: Step["tone"]): string {
  return `timeline-dot timeline-dot--${tone}`;
}
</script>

<template>
  <AppCard :title="t('compliance.exceptions.timeline.title')">
    <ol class="timeline">
      <li v-for="step in STEPS" :key="step.labelKey" class="timeline-item">
        <span :class="dotClass(step.tone)" aria-hidden="true" />
        <span class="timeline-label">{{ t(step.labelKey) }}</span>
      </li>
    </ol>
  </AppCard>
</template>

<style scoped>
.timeline {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-3);
}

.timeline-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.timeline-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--text-tertiary);
}

.timeline-dot--info { background: var(--state-info); }
.timeline-dot--success { background: var(--state-success); }
.timeline-dot--warning { background: var(--state-warning); }
.timeline-dot--danger { background: var(--state-danger); }
</style>
