<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import { useAuthStore } from "~/stores/useAuthStore";
import { useI18n } from "~/composables/useI18n";

import { myFundsApi } from "../../services/myFundsApi";
import { relativeTime } from "../../lib/format";
import type { ApiBreach } from "../../types";

interface Props {
  fundId: string;
}

const props = defineProps<Props>();
const { t } = useI18n();
const auth = useAuthStore();

const openBreaches = ref<ApiBreach[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const canRunCheck = computed(() => auth.hasPermission("INVESTMENT_LEDGER_SIMULATE"));
const canViewRules = computed(() => auth.hasPermission("IRG_VIEW_RULES"));

async function load() {
  loading.value = true;
  error.value = null;
  try {
    openBreaches.value = await myFundsApi.listOpenBreaches(props.fundId, 100);
  } catch (err) {
    openBreaches.value = [];
    error.value = describeError(err, "Failed to load breaches.");
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void load();
});

function describeError(err: unknown, fallback: string): string {
  if (err instanceof Error) return err.message;
  return fallback;
}

function severityVariant(severity: string | null | undefined): "danger" | "warn" | "info" | "muted" {
  const s = (severity ?? "").toUpperCase();
  if (s === "BLOCK") return "danger";
  if (s === "WARN") return "warn";
  if (s === "INFO") return "info";
  return "muted";
}
</script>

<template>
  <div class="compliance-panel">
    <AppCard
      :title="t('myFunds.detail.compliancePanel.title', 'Open compliance breaches')"
      :subtitle="
        t(
          'myFunds.detail.compliancePanel.subtitle',
          { count: openBreaches.length },
          `${openBreaches.length} open record(s)`,
        )
      "
    >
      <template #header-actions>
        <button
          type="button"
          class="compliance-panel__btn"
          @click="load"
          :disabled="loading"
        >
          {{ loading ? t("myFunds.detail.compliancePanel.refreshing", "Refreshing…") : t("myFunds.detail.compliancePanel.refresh", "Refresh") }}
        </button>
      </template>

      <div v-if="loading && openBreaches.length === 0" class="compliance-panel__loading">
        {{ t("myFunds.detail.compliancePanel.loading", "Loading breaches…") }}
      </div>

      <div v-else-if="error" class="compliance-panel__error" role="alert">
        {{ error }}
      </div>

      <div v-else-if="openBreaches.length === 0" class="compliance-panel__empty">
        {{ t("myFunds.detail.compliancePanel.empty", "No open breaches for this fund.") }}
      </div>

      <table v-else class="compliance-panel__table">
        <thead>
          <tr>
            <th>{{ t("myFunds.detail.compliancePanel.col.severity", "Severity") }}</th>
            <th>{{ t("myFunds.detail.compliancePanel.col.rule", "Rule") }}</th>
            <th>{{ t("myFunds.detail.compliancePanel.col.message", "Message") }}</th>
            <th>{{ t("myFunds.detail.compliancePanel.col.businessDate", "Business date") }}</th>
            <th>{{ t("myFunds.detail.compliancePanel.col.opened", "Opened") }}</th>
            <th>{{ t("myFunds.detail.compliancePanel.col.status", "Status") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="breach in openBreaches" :key="breach.id" :data-severity="severityVariant(breach.severity)">
            <td>
              <span class="compliance-panel__sev" :data-tone="severityVariant(breach.severity)">
                {{ breach.severity ?? "—" }}
              </span>
            </td>
            <td class="compliance-panel__rule">{{ breach.ruleTypeID ?? "—" }}</td>
            <td>{{ breach.message ?? "—" }}</td>
            <td class="compliance-panel__date">{{ breach.businessDate ?? "—" }}</td>
            <td class="compliance-panel__date">{{ relativeTime(breach.createdAt) }}</td>
            <td>{{ breach.status ?? "—" }}</td>
          </tr>
        </tbody>
      </table>
    </AppCard>

    <AppCard
      v-if="canRunCheck || canViewRules"
      :title="t('myFunds.detail.compliancePanel.actionsTitle', 'Compliance actions')"
    >
      <ul class="compliance-panel__action-list">
        <li v-if="canRunCheck">
          <router-link
            :to="{ path: '/compliance/pre-trade', query: { contract_id: fundId } }"
            class="compliance-panel__link"
          >
            {{ t("myFunds.detail.compliancePanel.runPreTrade", "Run pre-trade check") }}
          </router-link>
          <span class="compliance-panel__action-hint">
            {{
              t(
                "myFunds.detail.compliancePanel.runPreTradeHint",
                "Server-side gate; result is binding.",
              )
            }}
          </span>
        </li>
        <li v-if="canViewRules">
          <router-link to="/compliance/rules" class="compliance-panel__link">
            {{ t("myFunds.detail.compliancePanel.viewRules", "Browse IRG rules") }}
          </router-link>
        </li>
      </ul>
    </AppCard>
  </div>
</template>

<style scoped>
.compliance-panel {
  display: grid;
  gap: var(--space-3);
}

.compliance-panel__btn {
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  background: var(--bg-card, #ffffff);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 4px;
  padding: 4px 10px;
  cursor: pointer;
}

.compliance-panel__btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.compliance-panel__loading,
.compliance-panel__empty {
  padding: var(--space-3);
  font-size: 12px;
  color: var(--text-tertiary, #6e7781);
}

.compliance-panel__error {
  padding: var(--space-3);
  font-size: 12px;
  color: var(--state-danger, #cf222e);
}

.compliance-panel__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.compliance-panel__table thead th {
  text-align: left;
  font-weight: 600;
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  padding: 6px 8px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.compliance-panel__table tbody td {
  padding: 8px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  vertical-align: top;
  color: var(--text-primary, #1f2328);
}

.compliance-panel__sev {
  display: inline-block;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}

.compliance-panel__sev[data-tone="danger"] {
  background: rgba(207, 34, 46, 0.1);
  color: var(--state-danger, #cf222e);
}

.compliance-panel__sev[data-tone="warn"] {
  background: rgba(217, 119, 6, 0.12);
  color: var(--state-warning, #9a6700);
}

.compliance-panel__sev[data-tone="info"] {
  background: rgba(31, 111, 235, 0.12);
  color: var(--state-info, #1f6feb);
}

.compliance-panel__sev[data-tone="muted"] {
  background: var(--bg-card-muted, #f6f8fa);
  color: var(--text-tertiary, #6e7781);
}

.compliance-panel__rule,
.compliance-panel__date {
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 11px;
  color: var(--text-secondary, #57606a);
  white-space: nowrap;
}

.compliance-panel__action-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 8px;
}

.compliance-panel__link {
  font-weight: 600;
  color: var(--state-info, #1f6feb);
  text-decoration: none;
  font-size: 13px;
}

.compliance-panel__link:hover {
  text-decoration: underline;
}

.compliance-panel__action-hint {
  display: block;
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
  margin-top: 2px;
}
</style>
