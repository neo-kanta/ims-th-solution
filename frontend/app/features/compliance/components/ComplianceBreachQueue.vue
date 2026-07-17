<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";
import AppErrorState from "~/shared/ui/AppErrorState.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";
import AppSection from "~/shared/ui/AppSection.vue";

import {
  buildPortfolioComplianceHref,
  resolvePortfolioLink,
  type PortfolioLinkOption,
} from "../lib/breachQueue";
import { formatIsoDateTime } from "../lib/formatters";
import { ruleLabel } from "../lib/ruleTypeCatalog";
import type { ComplianceBreach } from "../types";

interface Props {
  items: ComplianceBreach[];
  total: number;
  loading: boolean;
  error: string | null;
  portfolioById: Map<string, PortfolioLinkOption>;
  limit?: number;
}

const props = withDefaults(defineProps<Props>(), { limit: 8 });
const emit = defineEmits<{ retry: [] }>();
const { t } = useI18n();

function severityBadgeClass(severity: string): string {
  if (severity === "BLOCK") {
    return "breach-row__severity breach-row__severity--block";
  }
  if (severity === "WARN" || severity === "REQUIRE_APPROVAL") {
    return "breach-row__severity breach-row__severity--warn";
  }
  return "breach-row__severity";
}

function severityLabel(severity: string): string {
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
      return severity;
  }
}

const rows = computed(() =>
  props.items.slice(0, props.limit).map((breach) => ({
    breach,
    portfolio: resolvePortfolioLink(
      breach.portfolioID,
      props.portfolioById,
      t("compliance.dashboard.breachQueue.portfolioUnavailable"),
    ),
  })),
);
</script>

<template>
  <AppSection
    :title="t('compliance.dashboard.breachQueue.title')"
    :description="t('compliance.dashboard.breachQueue.description')"
  >
    <AppLoadingState
      v-if="loading"
      skeleton
      :message="t('compliance.common.loading')"
    />

    <AppErrorState
      v-else-if="error"
      :title="t('compliance.dashboard.breachQueue.errorTitle')"
      :message="error"
      retry
      @retry="emit('retry')"
    />

    <AppEmptyState
      v-else-if="rows.length === 0"
      icon="alert"
      :title="t('compliance.dashboard.breachQueue.empty')"
      :description="t('compliance.dashboard.breachQueue.emptyDescription')"
    />

    <ul v-else class="breach-queue" role="list">
      <li v-for="row in rows" :key="row.breach.id" class="breach-row">
        <span :class="severityBadgeClass(row.breach.severity)">
          <span aria-hidden="true">{{ severityLabel(row.breach.severity) }}</span>
          <span class="sr-only">
            {{
              t("compliance.dashboard.breachQueue.severitySr", {
                severity: severityLabel(row.breach.severity),
              })
            }}
          </span>
        </span>

        <div class="breach-row__body">
          <div class="breach-row__rule">
            {{ ruleLabel(row.breach.ruleTypeID, t) }}
            <span class="breach-row__msg">{{ row.breach.message || "—" }}</span>
          </div>
          <div class="breach-row__meta">
            <NuxtLink
              v-if="row.portfolio.code"
              :to="buildPortfolioComplianceHref(row.portfolio.code)"
              class="breach-row__portfolio-link"
            >
              {{ row.portfolio.label }}
            </NuxtLink>
            <span v-else class="breach-row__portfolio-unavailable">
              {{ row.portfolio.label }}
            </span>
            <span aria-hidden="true">·</span>
            <span :title="formatIsoDateTime(row.breach.createdAt)">
              {{ formatIsoDateTime(row.breach.createdAt) }}
            </span>
          </div>
        </div>

        <NuxtLink
          v-if="row.portfolio.code"
          :to="buildPortfolioComplianceHref(row.portfolio.code)"
        >
          <AppButton variant="secondary" size="sm">
            {{ t("compliance.dashboard.breachQueue.review") }}
          </AppButton>
        </NuxtLink>
      </li>
    </ul>

    <div
      v-if="!loading && !error && total > rows.length"
      class="breach-queue__footer"
    >
      <p class="breach-queue__truncated">
        {{
          t("compliance.dashboard.breachQueue.truncated", {
            shown: rows.length,
            total,
          })
        }}
      </p>
      <NuxtLink to="/compliance/post-trade">
        {{ t("compliance.dashboard.headerActions.openBreaches") }}
      </NuxtLink>
    </div>
  </AppSection>
</template>

<style scoped>
.breach-queue {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
}

.breach-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-4) 0;
  border-bottom: 1px solid var(--border-subtle);
}

.breach-row:first-child {
  padding-top: 0;
}

.breach-row:last-child {
  padding-bottom: 0;
  border-bottom: 0;
}

.breach-row__severity {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0 var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-pill);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.05em;
  white-space: nowrap;
}

.breach-row__severity--block {
  border-color: var(--alert-danger-border);
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
}

.breach-row__severity--warn {
  border-color: var(--alert-warning-border);
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
}

.breach-row__body {
  min-width: 0;
  display: grid;
  gap: var(--space-1);
}

.breach-row__rule {
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.breach-row__msg {
  margin-left: var(--space-2);
  color: var(--text-secondary);
  font-weight: var(--font-weight-regular);
}

.breach-row__meta {
  display: inline-flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.breach-row__portfolio-link {
  color: var(--text-link);
  font-weight: var(--font-weight-medium);
  text-decoration: none;
}

.breach-row__portfolio-link:hover {
  text-decoration: underline;
}

.breach-row__portfolio-unavailable {
  font-style: italic;
}

.breach-queue__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding-top: var(--space-4);
}

.breach-queue__truncated {
  margin: 0;
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 640px) {
  .breach-row {
    grid-template-columns: 1fr;
    align-items: start;
  }

  .breach-row > a {
    justify-self: stretch;
  }

  .breach-row > a :deep(.btn) {
    width: 100%;
    justify-content: center;
  }

  .breach-queue__footer {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
