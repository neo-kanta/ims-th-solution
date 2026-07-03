<script setup lang="ts">
import { onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";

import { useCompliancePortfolioDirectory } from "../composables/useCompliancePortfolioDirectory";
import { formatIsoDate, formatIsoDateTime } from "../lib/formatters";
import { ruleLabel } from "../lib/ruleTypeCatalog";
import type { ComplianceBreach } from "../types";
import ComplianceSeverityBadge from "./ComplianceSeverityBadge.vue";
import ComplianceVerdictBadge from "./ComplianceVerdictBadge.vue";

interface Props {
  items: ComplianceBreach[];
  loading: boolean;
  error: string | null;
  canOverride: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  override: [breach: ComplianceBreach];
  "view-group": [groupId: string];
}>();

const { t } = useI18n();
const portfolios = useCompliancePortfolioDirectory();

onMounted(() => {
  void portfolios.ensureLoaded();
});

function portfolioLabel(id: string): string {
  return portfolios.labelFor(id);
}

function statusToneClass(status: string): string {
  switch (status) {
    case "OPEN":
      return "inbox__status inbox__status--open";
    case "OVERRIDDEN":
      return "inbox__status inbox__status--overridden";
    case "RESOLVED":
      return "inbox__status inbox__status--resolved";
    default:
      return "inbox__status";
  }
}

function shortenId(id: string, len = 8): string {
  if (!id) return "—";
  return id.length <= len ? id : `${id.slice(0, len)}…`;
}
</script>

<template>
  <div class="inbox-wrap">
    <div v-if="loading" class="inbox__state">
      {{ t("compliance.common.loading") }}
    </div>
    <div
      v-else-if="error"
      class="inbox__state inbox__state--error"
      role="alert"
    >
      {{ error }}
    </div>
    <AppEmptyState
      v-else-if="items.length === 0"
      icon="table"
      :title="t('compliance.postTrade.empty')"
    />
    <table v-else class="inbox-table">
      <thead>
        <tr>
          <th>{{ t("compliance.postTrade.table.rule") }}</th>
          <th>{{ t("compliance.postTrade.table.severity") }}</th>
          <th>{{ t("compliance.postTrade.table.verdict") }}</th>
          <th>{{ t("compliance.postTrade.table.status") }}</th>
          <th>{{ t("compliance.postTrade.table.businessDate") }}</th>
          <th>{{ t("compliance.postTrade.table.portfolio") }}</th>
          <th>{{ t("compliance.postTrade.table.contract") }}</th>
          <th>{{ t("compliance.postTrade.table.message") }}</th>
          <th class="inbox-table__actions">
            {{ t("compliance.postTrade.table.actions") }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="b in items" :key="b.id">
          <td>
            <div class="inbox__rule">{{ ruleLabel(b.ruleTypeID) }}</div>
            <code class="inbox__code">{{ b.ruleTypeID }}</code>
          </td>
          <td>
            <ComplianceSeverityBadge :severity="b.severity" />
          </td>
          <td>
            <ComplianceVerdictBadge :verdict="b.verdict" />
          </td>
          <td>
            <span :class="statusToneClass(b.status)">{{ b.status }}</span>
          </td>
          <td>{{ formatIsoDate(b.businessDate) }}</td>
          <td>
            <div class="inbox__portfolio">
              <span class="inbox__portfolio-label">
                {{ portfolioLabel(b.portfolioID) }}
              </span>
              <code class="inbox__code" :title="b.portfolioID">
                {{ shortenId(b.portfolioID) }}
              </code>
            </div>
          </td>
          <td>
            <code class="inbox__code" :title="b.contractID">
              {{ shortenId(b.contractID) }}
            </code>
          </td>
          <td class="inbox__msg" :title="b.message">{{ b.message || "—" }}</td>
          <td class="inbox-table__actions">
            <AppButton
              variant="primary"
              size="sm"
              :disabled="!canOverride || b.status !== 'OPEN'"
              :title="
                !canOverride
                  ? 'IRG_OVERRIDE_BREACH permission required.'
                  : b.status !== 'OPEN'
                    ? t('compliance.postTrade.override.notOverridable')
                    : undefined
              "
              @click="emit('override', b)"
            >
              {{ t("compliance.postTrade.table.override") }}
            </AppButton>
            <AppButton
              variant="ghost"
              size="sm"
              @click="emit('view-group', b.checkGroupID)"
            >
              {{ t("compliance.postTrade.table.viewGroup") }}
            </AppButton>
            <div class="inbox__transition-note">
              {{ t("compliance.postTrade.table.statusTransitionDisabled") }}
            </div>
            <div class="inbox__id-row">
              <span class="inbox__id-label">Breach</span>
              <code :title="b.id">{{ shortenId(b.id) }}</code>
              <span class="inbox__id-label">·</span>
              <span>{{ formatIsoDateTime(b.createdAt) }}</span>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.inbox-wrap {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  overflow-x: auto;
}

.inbox__state {
  padding: var(--space-6);
  color: var(--text-secondary);
}

.inbox__state--error {
  color: var(--state-danger);
}

.inbox-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
}

.inbox-table thead th {
  background: var(--bg-table-header);
  color: var(--text-secondary);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-xs);
  text-align: left;
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  white-space: nowrap;
}

.inbox-table tbody td {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
  vertical-align: top;
}

.inbox__rule {
  font-weight: var(--font-weight-semibold);
}

.inbox__code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.inbox__portfolio {
  display: grid;
  gap: var(--space-1);
}

.inbox__portfolio-label {
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.inbox__msg {
  max-width: 22rem;
  color: var(--text-secondary);
}

.inbox-table__actions {
  white-space: nowrap;
  display: grid;
  gap: var(--space-2);
  align-items: start;
}

.inbox-table__actions {
  min-width: 220px;
}

.inbox__transition-note {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  font-style: italic;
}

.inbox__id-row {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.inbox__id-label {
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.inbox__status {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 var(--space-3);
  border-radius: var(--radius-pill);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  border: 1px solid var(--border-subtle);
}

.inbox__status--open {
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border-color: var(--alert-danger-border);
}

.inbox__status--overridden {
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  border-color: var(--alert-warning-border);
}

.inbox__status--resolved {
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
  border-color: var(--alert-success-border);
}
</style>
