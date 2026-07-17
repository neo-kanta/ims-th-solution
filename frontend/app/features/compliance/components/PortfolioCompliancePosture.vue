<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppBadge from "~/shared/ui/AppBadge.vue";

import ComplianceKpiCard from "./ComplianceKpiCard.vue";

interface Props {
  openBreachCount: number | null;
  effectiveCount: number | null;
  scheduledCount: number | null;
  strongestEffectiveSeverity: string | null;
  loading: boolean;
}

const props = defineProps<Props>();

const { t } = useI18n();

const severityVariant = computed(() => {
  switch (props.strongestEffectiveSeverity) {
    case "BLOCK":
      return "error" as const;
    case "WARN":
    case "REQUIRE_APPROVAL":
      return "warning" as const;
    case "MONITOR":
      return "neutral" as const;
    default:
      return "neutral" as const;
  }
});

const severityLabel = computed(() => {
  switch (props.strongestEffectiveSeverity) {
    case "BLOCK":
      return t("compliance.badges.severity.BLOCK");
    case "WARN":
      return t("compliance.badges.severity.WARN");
    case "REQUIRE_APPROVAL":
      return t("compliance.badges.severity.REQUIRE_APPROVAL");
    case "MONITOR":
      return t("compliance.badges.severity.MONITOR");
    default:
      return t("portfolio.compliance.posture.noneEffective");
  }
});
</script>

<template>
  <section class="posture" :aria-label="t('portfolio.compliance.posture.groupLabel')">
    <ComplianceKpiCard
      :label="t('portfolio.compliance.posture.openBreaches')"
      :value="openBreachCount"
      tone="danger"
      :loading="loading"
    />
    <ComplianceKpiCard
      :label="t('portfolio.compliance.posture.effective')"
      :value="effectiveCount"
      tone="success"
      :loading="loading"
    />
    <ComplianceKpiCard
      :label="t('portfolio.compliance.posture.scheduled')"
      :value="scheduledCount"
      tone="default"
      :loading="loading"
    />
    <article class="posture__severity" :aria-busy="loading">
      <span class="posture__severity-label">{{ t("portfolio.compliance.posture.strongestSeverity") }}</span>
      <div class="posture__severity-value">
        <template v-if="loading">…</template>
        <AppBadge v-else :variant="severityVariant" dot>{{ severityLabel }}</AppBadge>
      </div>
    </article>
  </section>
</template>

<style scoped>
.posture {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
}

@media (max-width: 900px) {
  .posture {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 480px) {
  .posture {
    grid-template-columns: 1fr;
  }
}

.posture__severity {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  min-width: 0;
}

.posture__severity-label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.posture__severity-value {
  font-size: var(--font-size-lg);
}
</style>
