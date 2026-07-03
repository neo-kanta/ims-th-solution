<script setup lang="ts">
import { computed, onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";

import { useComplianceUserDirectory } from "../composables/useComplianceUserDirectory";
import {
  deriveRuleStatus,
  formatIsoDate,
  formatIsoDateTime,
  formatEffectiveWindow,
} from "../lib/formatters";
import { ruleLabel } from "../lib/ruleTypeCatalog";
import type { ComplianceRule } from "../types";
import ComplianceRuleStatusBadge from "./ComplianceRuleStatusBadge.vue";
import ComplianceSeverityBadge from "./ComplianceSeverityBadge.vue";

interface Props {
  items: ComplianceRule[];
  loading: boolean;
  error: string | null;
  highlightId?: string | null;
}

const props = withDefaults(defineProps<Props>(), {
  highlightId: null,
});

const { t } = useI18n();
const users = useComplianceUserDirectory();

onMounted(() => {
  void users.ensureLoaded();
});

const rows = computed(() =>
  props.items.map((rule) => ({
    rule,
    derivedStatus: deriveRuleStatus(rule),
    label: ruleLabel(rule.ruleTypeID),
  })),
);

function truncate(id: string, len = 8): string {
  if (!id) return "—";
  return id.length <= len ? id : `${id.slice(0, len)}…`;
}
</script>

<template>
  <div class="rules-table-wrap">
    <div v-if="loading" class="rules-table__state">
      {{ t("compliance.common.loading") }}
    </div>
    <div
      v-else-if="error"
      class="rules-table__state rules-table__state--error"
      role="alert"
    >
      {{ error }}
    </div>
    <AppEmptyState
      v-else-if="items.length === 0"
      icon="table"
      :title="t('compliance.rules.table.empty')"
    />
    <table v-else class="rules-table">
      <thead>
        <tr>
          <th>{{ t("compliance.rules.table.code") }}</th>
          <th>{{ t("compliance.rules.table.name") }}</th>
          <th>{{ t("compliance.rules.table.category") }}</th>
          <th>{{ t("compliance.rules.table.severity") }}</th>
          <th>{{ t("compliance.rules.table.status") }}</th>
          <th>{{ t("compliance.rules.table.version") }}</th>
          <th>{{ t("compliance.rules.table.effective") }}</th>
          <th>{{ t("compliance.rules.table.owner") }}</th>
          <th>{{ t("compliance.rules.table.updated") }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="row in rows"
          :key="row.rule.id"
          :class="{
            'rules-table__row--highlight': highlightId === row.rule.id,
          }"
        >
          <td class="rules-table__code">{{ row.rule.ruleTypeID }}</td>
          <td>
            <NuxtLink
              :to="`/compliance/rules/${row.rule.id}`"
              class="rules-table__name-link"
            >
              <div class="rules-table__name">
                {{ row.rule.name || row.label }}
              </div>
            </NuxtLink>
            <div v-if="row.rule.description" class="rules-table__desc">
              {{ row.rule.description }}
            </div>
          </td>
          <td>{{ row.rule.type_metadata?.category ?? "—" }}</td>
          <td>
            <ComplianceSeverityBadge
              v-if="row.rule.type_metadata?.default_severity"
              :severity="row.rule.type_metadata.default_severity"
            />
            <span v-else class="rules-table__muted">—</span>
          </td>
          <td>
            <ComplianceRuleStatusBadge :status="row.derivedStatus" />
          </td>
          <td class="rules-table__num">{{ row.rule.currentVersion }}</td>
          <td>{{ formatEffectiveWindow(row.rule.effectiveWindow) }}</td>
          <td>
            <div class="rules-table__owner-cell" :title="row.rule.createdBy">
              <span>{{ users.labelFor(row.rule.createdBy) }}</span>
              <code v-if="!users.forbidden.value" class="rules-table__owner">
                {{ truncate(row.rule.createdBy) }}
              </code>
            </div>
          </td>
          <td class="rules-table__updated">
            {{ formatIsoDateTime(row.rule.updatedAt) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.rules-table-wrap {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  overflow-x: auto;
}

.rules-table__state {
  padding: var(--space-6);
  color: var(--text-secondary);
}

.rules-table__state--error {
  color: var(--state-danger);
}

.rules-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
}

.rules-table thead th {
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

.rules-table tbody td {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
  vertical-align: top;
}

.rules-table tbody tr:hover {
  background: var(--bg-row-hover);
}

.rules-table__row--highlight {
  background: var(--bg-selected) !important;
}

.rules-table__code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  white-space: nowrap;
}

.rules-table__name {
  font-weight: var(--font-weight-semibold);
}

.rules-table__name-link {
  color: var(--text-link);
  text-decoration: none;
}

.rules-table__name-link:hover {
  text-decoration: underline;
}

.rules-table__desc {
  margin-top: var(--space-1);
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  max-width: 28rem;
}

.rules-table__num {
  font-variant-numeric: tabular-nums;
  color: var(--text-secondary);
}

.rules-table__owner-cell {
  display: grid;
  gap: 2px;
}

.rules-table__owner {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.rules-table__muted {
  color: var(--text-tertiary);
}

.rules-table__updated {
  white-space: nowrap;
  color: var(--text-secondary);
}
</style>
