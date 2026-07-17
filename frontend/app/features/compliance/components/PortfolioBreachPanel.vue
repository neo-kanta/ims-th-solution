<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";
import AppErrorState from "~/shared/ui/AppErrorState.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";
import AppSection from "~/shared/ui/AppSection.vue";

import { sortPortfolioBreaches } from "../lib/portfolioPosture";
import { ruleLabel } from "../lib/ruleTypeCatalog";
import { formatIsoDate, formatIsoDateTime } from "../lib/formatters";
import type { ApiPortfolioBreachView } from "../../portfolio-workspace/services/portfolioComplianceApi";

interface Props {
  breaches: ApiPortfolioBreachView[];
  loading: boolean;
  error: string | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{ retry: [] }>();

const { t } = useI18n();

const openBreaches = computed(() =>
  sortPortfolioBreaches(props.breaches.filter((b) => b.status === "OPEN")),
);
const resolvedCount = computed(
  () => props.breaches.length - openBreaches.value.length,
);

function severityLabel(severity: string | undefined): string {
  switch (severity) {
    case "BLOCK":
      return t("compliance.badges.severity.BLOCK");
    case "WARN":
      return t("compliance.badges.severity.WARN");
    case "REQUIRE_APPROVAL":
      return t("compliance.badges.severity.REQUIRE_APPROVAL");
    case "MONITOR":
      return t("compliance.badges.severity.MONITOR");
    default:
      return severity || "—";
  }
}

function severityClass(severity: string | undefined): string {
  if (severity === "BLOCK") return "breach__severity breach__severity--block";
  if (severity === "WARN" || severity === "REQUIRE_APPROVAL") return "breach__severity breach__severity--warn";
  return "breach__severity";
}
</script>

<template>
  <AppSection
    :title="t('portfolio.compliance.breaches.title')"
    :description="t('portfolio.compliance.breaches.description')"
  >
    <AppLoadingState v-if="loading" skeleton :message="t('compliance.common.loading')" />

    <AppErrorState
      v-else-if="error"
      :title="t('portfolio.compliance.breaches.errorTitle')"
      :message="error"
      retry
      @retry="emit('retry')"
    />

    <AppEmptyState
      v-else-if="openBreaches.length === 0"
      icon="alert"
      :title="t('portfolio.compliance.breaches.empty')"
    />

    <ul v-else class="breach-list" role="list">
      <li v-for="(b, idx) in openBreaches" :key="`${b.rule_type_id}-${b.created_at}-${idx}`" class="breach">
        <span :class="severityClass(b.severity)">{{ severityLabel(b.severity) }}</span>
        <div class="breach__body">
          <div class="breach__rule">{{ ruleLabel(b.rule_type_id ?? "", t) }}</div>
          <div class="breach__msg">{{ b.message || "—" }}</div>
          <div class="breach__meta">
            <span>{{ t("portfolio.compliance.breaches.businessDate", { date: formatIsoDate(b.business_date) }) }}</span>
            <span aria-hidden="true">·</span>
            <span :title="formatIsoDateTime(b.created_at)">{{ formatIsoDateTime(b.created_at) }}</span>
          </div>
        </div>
      </li>
    </ul>

    <p v-if="!loading && !error && resolvedCount > 0" class="breach-list__resolved-note">
      {{ t("portfolio.compliance.breaches.resolvedNote", { count: resolvedCount }) }}
    </p>
  </AppSection>
</template>

<style scoped>
.breach-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-3);
}

.breach {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.breach__severity {
  display: inline-flex;
  align-items: center;
  height: 24px;
  padding: 0 var(--space-3);
  border-radius: var(--radius-pill);
  background: var(--bg-card-muted);
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.05em;
  white-space: nowrap;
  align-self: start;
}

.breach__severity--block {
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border-color: var(--alert-danger-border);
}

.breach__severity--warn {
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  border-color: var(--alert-warning-border);
}

.breach__body {
  min-width: 0;
  display: grid;
  gap: 2px;
}

.breach__rule {
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.breach__msg {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.breach__meta {
  display: inline-flex;
  gap: var(--space-2);
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.breach-list__resolved-note {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}
</style>
