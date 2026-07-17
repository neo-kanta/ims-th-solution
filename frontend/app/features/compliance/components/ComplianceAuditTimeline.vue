<script setup lang="ts">
import { computed, onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppCard from "~/shared/ui/AppCard.vue";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";

import { useCompliancePortfolioDirectory } from "../composables/useCompliancePortfolioDirectory";
import { useComplianceUserDirectory } from "../composables/useComplianceUserDirectory";
import { formatIsoDate, formatIsoDateTime } from "../lib/formatters";
import { ruleLabel } from "../lib/ruleTypeCatalog";
import type { ComplianceCheckGroupResult } from "../types";
import ComplianceSeverityBadge from "./ComplianceSeverityBadge.vue";
import ComplianceVerdictBadge from "./ComplianceVerdictBadge.vue";

interface Props {
  data: ComplianceCheckGroupResult | null;
  loading: boolean;
  error: string | null;
}

const props = defineProps<Props>();

const { t } = useI18n();
const portfolios = useCompliancePortfolioDirectory();
const users = useComplianceUserDirectory();

onMounted(() => {
  void Promise.all([portfolios.ensureLoaded(), users.ensureLoaded()]);
});

const recordCount = computed(() => props.data?.records?.length ?? 0);
const breachCount = computed(() => props.data?.breaches?.length ?? 0);

function shortenId(id: string | null | undefined, len = 8): string {
  if (!id) return t("common.notAvailable");
  return id.length <= len ? id : `${id.slice(0, len)}…`;
}

function userLabel(id: string | null | undefined): string {
  return users.labelFor(id) || t("common.notAvailable");
}
</script>

<template>
  <div class="audit-timeline">
    <div v-if="loading" class="audit-timeline__state">
      {{ t("compliance.audit.lookup.loading") }}
    </div>
    <div
      v-else-if="error"
      class="audit-timeline__state audit-timeline__state--error"
      role="alert"
    >
      {{ error }}
    </div>
    <AppEmptyState
      v-else-if="!data"
      icon="search"
      :title="t('compliance.audit.lookup.empty')"
    />
    <template v-else>
      <AppCard
        :title="t('compliance.audit.records') + ` (${recordCount})`"
      >
        <AppEmptyState
          v-if="recordCount === 0"
          icon="table"
          :title="t('compliance.postTrade.groupDrawer.empty')"
        />
        <ol v-else class="audit-list">
          <li
            v-for="r in data.records"
            :key="r.id"
            class="audit-item"
            :class="{
              'audit-item--block': r.finalVerdict === 'BLOCK',
              'audit-item--warn': r.finalVerdict === 'WARN',
            }"
          >
            <div class="audit-item__head">
              <div class="audit-item__title">{{ ruleLabel(r.ruleTypeID, t) }}</div>
              <div class="audit-item__badges">
                <ComplianceVerdictBadge :verdict="r.finalVerdict" />
                <ComplianceSeverityBadge :severity="r.effectiveSeverity" />
              </div>
            </div>
            <p class="audit-item__message">{{ r.message || t("common.notAvailable") }}</p>
            <dl class="audit-item__meta">
              <div>
                <dt>{{ t("compliance.audit.fields.timing") }}</dt>
                <dd>{{ r.timing }}</dd>
              </div>
              <div>
                <dt>{{ t("compliance.audit.fields.businessDate") }}</dt>
                <dd>{{ formatIsoDate(r.businessDate) }}</dd>
              </div>
              <div>
                <dt>{{ t("compliance.audit.fields.checkedAt") }}</dt>
                <dd>{{ formatIsoDateTime(r.checkedAt) }}</dd>
              </div>
              <div>
                <dt>{{ t("compliance.audit.fields.checkedBy") }}</dt>
                <dd>{{ userLabel(r.checkedBy ?? null) }}</dd>
              </div>
              <div>
                <dt>{{ t("compliance.audit.fields.order") }}</dt>
                <dd><code>{{ shortenId(r.orderID ?? null) }}</code></dd>
              </div>
              <div>
                <dt>{{ t("compliance.audit.fields.ticker") }}</dt>
                <dd>{{ r.ticker || t("common.notAvailable") }}</dd>
              </div>
              <div>
                <dt>{{ t("compliance.audit.fields.portfolio") }}</dt>
                <dd>{{ portfolios.labelFor(r.portfolioID) || t("common.notAvailable") }}</dd>
              </div>
              <div>
                <dt>{{ t("compliance.audit.fields.ruleVersion") }}</dt>
                <dd>v{{ r.ruleInstanceVersion }}</dd>
              </div>
              <div>
                <dt>{{ t("compliance.audit.fields.dataHash") }}</dt>
                <dd><code>{{ shortenId(r.dataSnapshotHash ?? null, 12) }}</code></dd>
              </div>
            </dl>
          </li>
        </ol>
      </AppCard>

      <AppCard :title="t('compliance.audit.breaches') + ` (${breachCount})`">
        <AppEmptyState
          v-if="breachCount === 0"
          icon="table"
          :title="t('compliance.audit.noBreaches')"
        />
        <ol v-else class="audit-list">
          <li
            v-for="b in data.breaches"
            :key="b.id"
            class="audit-item"
            :class="{
              'audit-item--block': b.verdict === 'BLOCK',
              'audit-item--warn': b.verdict === 'WARN',
            }"
          >
            <div class="audit-item__head">
              <div class="audit-item__title">{{ ruleLabel(b.ruleTypeID, t) }}</div>
              <div class="audit-item__badges">
                <ComplianceVerdictBadge :verdict="b.verdict" />
                <ComplianceSeverityBadge :severity="b.severity" />
                <span class="audit-item__status">{{ b.status }}</span>
              </div>
            </div>
            <p class="audit-item__message">{{ b.message || t("common.notAvailable") }}</p>
            <dl class="audit-item__meta">
              <div>
                <dt>{{ t("compliance.audit.fields.breach") }}</dt>
                <dd><code>{{ b.id }}</code></dd>
              </div>
              <div>
                <dt>{{ t("compliance.audit.fields.businessDate") }}</dt>
                <dd>{{ formatIsoDate(b.businessDate) }}</dd>
              </div>
              <div>
                <dt>{{ t("compliance.audit.fields.created") }}</dt>
                <dd>{{ formatIsoDateTime(b.createdAt) }}</dd>
              </div>
            </dl>
          </li>
        </ol>
      </AppCard>
    </template>
  </div>
</template>

<style scoped>
.audit-timeline {
  display: grid;
  gap: var(--space-5);
}

.audit-timeline__state {
  padding: var(--space-5);
  color: var(--text-secondary);
}

.audit-timeline__state--error {
  color: var(--state-danger);
}

.audit-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-4);
}

.audit-item {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  border-left: 4px solid var(--border-subtle);
}

.audit-item--block {
  border-left-color: var(--state-danger);
}

.audit-item--warn {
  border-left-color: var(--state-warning);
}

.audit-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.audit-item__title {
  font-weight: var(--font-weight-semibold);
}

.audit-item__badges {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.audit-item__status {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 var(--space-3);
  border-radius: var(--radius-pill);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
}

.audit-item__message {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.audit-item__meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: var(--space-3);
  margin: 0;
}

.audit-item__meta dt {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.audit-item__meta dd {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.audit-item__meta code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}
</style>
