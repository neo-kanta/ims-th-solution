<script setup lang="ts">
/**
 * Compact results queue for the Breaches page — a dense desktop table plus a
 * stacked mobile list, both driven by the same `items`. Replaces the former
 * nine-column table with permanent full-width action buttons.
 */
import { onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";
import AppErrorState from "~/shared/ui/AppErrorState.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";

import { useBangkokFormatter } from "~/shared/composables/useBangkokFormatter";
import { useCompliancePortfolioDirectory } from "../composables/useCompliancePortfolioDirectory";
import { canOverrideBreach, overrideBlockedReason } from "../lib/breachActions";
import { formatIsoDate } from "../lib/formatters";
import { ruleLabel } from "../lib/ruleTypeCatalog";
import type { ComplianceBreach, ComplianceBreachStatus } from "../types";
import ComplianceSeverityBadge from "./ComplianceSeverityBadge.vue";
import ComplianceVerdictBadge from "./ComplianceVerdictBadge.vue";

interface Props {
  items: ComplianceBreach[];
  loading: boolean;
  error: string | null;
  canOverride: boolean;
  hasActiveFilters: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  override: [breach: ComplianceBreach];
  "view-details": [breach: ComplianceBreach];
  retry: [];
  "clear-filters": [];
}>();

const { t } = useI18n();
const { formatDateTime } = useBangkokFormatter();
const portfolios = useCompliancePortfolioDirectory();

onMounted(() => {
  void portfolios.ensureLoaded();
});

function portfolioCode(id: string): string {
  const hit = portfolios.byId.value.get(id);
  return hit ? hit.code : t("common.notAvailable");
}

function portfolioName(id: string): string | null {
  const hit = portfolios.byId.value.get(id);
  return hit ? hit.name : null;
}

function statusTone(status: ComplianceBreachStatus): "active" | "warning" | "success" {
  switch (status) {
    case "OPEN":
      return "active";
    case "OVERRIDDEN":
      return "warning";
    case "RESOLVED":
      return "success";
    default:
      return "active";
  }
}

function statusLabel(status: ComplianceBreachStatus): string {
  switch (status) {
    case "OPEN":
      return t("compliance.postTrade.toolbar.statusOpen");
    case "OVERRIDDEN":
      return t("compliance.postTrade.toolbar.statusOverridden");
    case "RESOLVED":
      return t("compliance.postTrade.toolbar.statusResolved");
    default:
      return status;
  }
}

function overridable(breach: ComplianceBreach): boolean {
  return canOverrideBreach(breach, props.canOverride);
}

function overrideTitle(breach: ComplianceBreach): string | undefined {
  const reason = overrideBlockedReason(breach, props.canOverride);
  if (reason === "NOT_OPEN") return t("compliance.postTrade.table.overrideNotOpen");
  if (reason === "NO_PERMISSION") return t("compliance.postTrade.table.overridePermissionRequired");
  return undefined;
}
</script>

<template>
  <div class="breach-queue">
    <AppLoadingState
      v-if="loading && items.length === 0"
      skeleton
      :message="t('compliance.postTrade.states.loading')"
    />

    <AppErrorState
      v-else-if="error && items.length === 0"
      :title="t('compliance.postTrade.states.errorTitle')"
      :message="error"
      retry
      @retry="emit('retry')"
    />

    <AppEmptyState
      v-else-if="items.length === 0"
      icon="table"
      :title="
        hasActiveFilters
          ? t('compliance.postTrade.states.emptyFilteredTitle')
          : t('compliance.postTrade.states.emptyNoBreachesTitle')
      "
      :description="
        hasActiveFilters
          ? t('compliance.postTrade.states.emptyFilteredDescription')
          : t('compliance.postTrade.states.emptyNoBreachesDescription')
      "
    >
      <template v-if="hasActiveFilters" #action>
        <AppButton variant="secondary" size="sm" @click="emit('clear-filters')">
          {{ t("compliance.postTrade.states.clearFilters") }}
        </AppButton>
      </template>
    </AppEmptyState>

    <template v-else>
      <p v-if="loading" class="breach-queue__refreshing" role="status">
        {{ t("compliance.postTrade.refreshing") }}
      </p>
      <p v-if="error" class="breach-queue__inline-error" role="alert">{{ error }}</p>

      <!-- Desktop / tablet table -->
      <table class="breach-table">
        <thead>
          <tr>
            <th class="breach-table__breach-col">{{ t("compliance.postTrade.table.breach") }}</th>
            <th>{{ t("compliance.postTrade.table.risk") }}</th>
            <th>{{ t("compliance.postTrade.table.portfolio") }}</th>
            <th>{{ t("compliance.postTrade.table.businessDate") }}</th>
            <th>{{ t("compliance.postTrade.table.status") }}</th>
            <th class="breach-table__actions-col">{{ t("compliance.postTrade.table.actions") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in items" :key="b.id">
            <td class="breach-table__breach-cell">
              <button
                type="button"
                class="breach-link"
                :title="b.message || undefined"
                @click="emit('view-details', b)"
              >
                <span class="breach-link__rule">{{ ruleLabel(b.ruleTypeID, t) }}</span>
                <span class="breach-link__message">{{
                  b.message || t("compliance.postTrade.drawer.messageEmpty")
                }}</span>
              </button>
            </td>
            <td>
              <div class="breach-table__risk">
                <ComplianceSeverityBadge :severity="b.severity" />
                <ComplianceVerdictBadge :verdict="b.verdict" />
              </div>
            </td>
            <td>
              <div class="breach-table__portfolio">
                <span class="breach-table__portfolio-code">{{ portfolioCode(b.portfolioID) }}</span>
                <span v-if="portfolioName(b.portfolioID)" class="breach-table__portfolio-name">
                  {{ portfolioName(b.portfolioID) }}
                </span>
              </div>
            </td>
            <td>
              <div class="breach-table__dates">
                <span>{{ formatIsoDate(b.businessDate) }}</span>
                <span class="breach-table__created">{{ formatDateTime(b.createdAt) }}</span>
              </div>
            </td>
            <td>
              <AppStatusBadge :status="statusTone(b.status)" :label="statusLabel(b.status)" size="sm" />
            </td>
            <td class="breach-table__actions-cell">
              <AppButton
                v-if="overridable(b)"
                variant="primary"
                size="sm"
                @click="emit('override', b)"
              >
                {{ t("compliance.postTrade.table.override") }}
              </AppButton>
              <span
                v-else-if="b.status === 'OPEN'"
                class="breach-table__action-note"
                :title="overrideTitle(b)"
              >
                {{ t("compliance.postTrade.table.noAction") }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Mobile stacked queue -->
      <ul class="breach-list">
        <li v-for="b in items" :key="b.id" class="breach-list__item">
          <button
            type="button"
            class="breach-list__link"
            @click="emit('view-details', b)"
          >
            <div class="breach-list__top">
              <span class="breach-list__rule">{{ ruleLabel(b.ruleTypeID, t) }}</span>
              <div class="breach-list__risk">
                <ComplianceSeverityBadge :severity="b.severity" size="sm" />
                <ComplianceVerdictBadge :verdict="b.verdict" size="sm" />
              </div>
            </div>
            <p class="breach-list__message">
              {{ b.message || t("compliance.postTrade.drawer.messageEmpty") }}
            </p>
            <div class="breach-list__meta">
              <span>{{ portfolioCode(b.portfolioID) }}</span>
              <span aria-hidden="true">·</span>
              <span>{{ formatIsoDate(b.businessDate) }}</span>
            </div>
          </button>
          <div class="breach-list__footer">
            <AppStatusBadge :status="statusTone(b.status)" :label="statusLabel(b.status)" size="sm" />
            <AppButton
              v-if="overridable(b)"
              variant="primary"
              size="sm"
              @click="emit('override', b)"
            >
              {{ t("compliance.postTrade.table.override") }}
            </AppButton>
            <AppButton v-else variant="ghost" size="sm" @click="emit('view-details', b)">
              {{ t("compliance.postTrade.table.viewDetails") }}
            </AppButton>
          </div>
        </li>
      </ul>
    </template>
  </div>
</template>

<style scoped>
.breach-queue {
  display: grid;
  gap: var(--space-3);
}

.breach-queue__refreshing {
  margin: 0;
  padding: var(--space-2) var(--space-3);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.breach-queue__inline-error {
  margin: 0;
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--alert-danger-border);
  border-radius: var(--radius-md);
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  font-size: var(--font-size-xs);
}

/* Desktop table */
.breach-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.breach-table thead th {
  background: var(--bg-table-header);
  color: var(--text-secondary);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-xs);
  text-align: left;
  padding: var(--space-2) var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  white-space: nowrap;
}

.breach-table__breach-col {
  width: 34%;
}

.breach-table__actions-col {
  width: 9rem;
}

.breach-table tbody tr {
  height: 68px;
}

.breach-table tbody tr:not(:last-child) td {
  border-bottom: 1px solid var(--border-subtle);
}

.breach-table tbody td {
  padding: var(--space-2) var(--space-4);
  color: var(--text-primary);
  vertical-align: middle;
}

.breach-link {
  display: grid;
  gap: 2px;
  width: 100%;
  max-width: 26rem;
  border: none;
  background: transparent;
  padding: 0;
  text-align: left;
  cursor: pointer;
  color: inherit;
  font: inherit;
}

.breach-link:focus-visible {
  outline: 2px solid var(--border-focus);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

.breach-link__rule {
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.breach-link__message {
  overflow: hidden;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.breach-table__risk {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
}

.breach-table__portfolio {
  display: grid;
  gap: 1px;
  max-width: 12rem;
}

.breach-table__portfolio-code {
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
}

.breach-table__portfolio-name {
  overflow: hidden;
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.breach-table__dates {
  display: grid;
  gap: 1px;
  font-variant-numeric: tabular-nums;
}

.breach-table__created {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.breach-table__actions-cell {
  text-align: right;
}

.breach-table__action-note {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

/* Mobile stacked list */
.breach-list {
  display: none;
  margin: 0;
  padding: 0;
  list-style: none;
  gap: var(--space-2);
}

.breach-list__item {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
}

.breach-list__link {
  display: grid;
  gap: var(--space-1);
  width: 100%;
  border: none;
  background: transparent;
  padding: 0;
  text-align: left;
  cursor: pointer;
  color: inherit;
  font: inherit;
}

.breach-list__link:focus-visible {
  outline: 2px solid var(--border-focus);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

.breach-list__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
}

.breach-list__rule {
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.breach-list__risk {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
  flex-shrink: 0;
}

.breach-list__message {
  margin: 0;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.breach-list__meta {
  display: flex;
  gap: var(--space-1);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.breach-list__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

@media (max-width: 768px) {
  .breach-table {
    display: none;
  }

  .breach-list {
    display: grid;
  }
}

@media (prefers-reduced-motion: reduce) {
  .breach-link,
  .breach-list__link {
    transition: none;
  }
}
</style>
