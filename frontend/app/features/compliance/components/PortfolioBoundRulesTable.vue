<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppBadge from "~/shared/ui/AppBadge.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppDataTable from "~/shared/ui/AppDataTable.vue";
import type { TableColumn } from "~/shared/ui/AppDataTable.vue";
import AppSection from "~/shared/ui/AppSection.vue";

import { deriveBindingState } from "../../portfolio-workspace/lib/complianceBindingState";
import type { ApiPortfolioRuleCatalogEntry } from "../../portfolio-workspace/services/portfolioComplianceApi";
import { formatIsoDate } from "../lib/formatters";
import { formatParameterEntries } from "../lib/ruleParameters";
import { ruleExplanation, ruleLabel } from "../lib/ruleTypeCatalog";

interface Props {
  entries: ApiPortfolioRuleCatalogEntry[];
  todayIso: string;
  busy: Record<string, boolean>;
  canManage: boolean;
  loading: boolean;
  error: string | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  deactivate: [entry: ApiPortfolioRuleCatalogEntry];
}>();

const { t } = useI18n();

const columns = computed<TableColumn[]>(() => [
  { key: "rule", label: t("portfolio.compliance.columns.rule") },
  { key: "parameters", label: t("portfolio.compliance.columns.parameters") },
  { key: "severity", label: t("portfolio.compliance.columns.severity") },
  { key: "effective", label: t("portfolio.compliance.columns.effective") },
  { key: "status", label: t("portfolio.compliance.columns.status") },
  { key: "actions", label: t("portfolio.compliance.columns.actions"), align: "right" },
]);

const rows = computed(() =>
  props.entries.map((entry) => ({ id: entry.rule_instance_id ?? "", entry })),
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

function severityVariant(severity: string | undefined) {
  if (severity === "BLOCK") return "error" as const;
  if (severity === "WARN" || severity === "REQUIRE_APPROVAL") return "warning" as const;
  if (severity === "MONITOR") return "neutral" as const;
  return "neutral" as const;
}

function stateLabel(entry: ApiPortfolioRuleCatalogEntry): string {
  const state = deriveBindingState(entry.binding, props.todayIso);
  switch (state) {
    case "SCHEDULED":
      return t("portfolio.compliance.state.scheduled");
    case "EFFECTIVE":
      return t("portfolio.compliance.state.effective");
    case "EXPIRED":
      return t("portfolio.compliance.state.expired");
    case "DEACTIVATED":
      return t("portfolio.compliance.state.deactivated");
    default:
      return "—";
  }
}

function stateVariant(entry: ApiPortfolioRuleCatalogEntry) {
  const state = deriveBindingState(entry.binding, props.todayIso);
  if (state === "EFFECTIVE") return "success" as const;
  if (state === "SCHEDULED") return "info" as const;
  return "neutral" as const;
}

function effectiveWindowLabel(entry: ApiPortfolioRuleCatalogEntry): string {
  const from = formatIsoDate(entry.binding?.effective_from);
  const to = entry.binding?.effective_to
    ? formatIsoDate(entry.binding.effective_to)
    : t("portfolio.compliance.indefinite");
  return `${from} – ${to}`;
}
</script>

<template>
  <AppSection
    :title="t('portfolio.compliance.boundTitle')"
    :description="t('portfolio.compliance.boundSubtitle')"
  >
  <AppDataTable
    :columns="columns"
    :items="rows"
    :loading="loading"
    :error="error"
    :empty-text="t('portfolio.compliance.noneBound')"
  >
    <template #cell(rule)="{ item }">
      <div class="rule-cell__label">{{ ruleLabel(item.entry.rule_type_id ?? "", t) }}</div>
      <div class="rule-cell__explanation">{{ ruleExplanation(item.entry.rule_type_id ?? "", item.entry.description ?? "", t) }}</div>
      <code class="rule-cell__type-id">{{ item.entry.rule_type_id }}</code>
    </template>

    <template #cell(parameters)="{ item }">
      <dl v-if="formatParameterEntries(item.entry.parameters).length > 0" class="param-list">
        <div v-for="p in formatParameterEntries(item.entry.parameters)" :key="p.key" class="param-list__row">
          <dt>{{ p.label }}</dt>
          <dd>{{ p.value }}</dd>
        </div>
      </dl>
      <span v-else>—</span>
    </template>

    <template #cell(severity)="{ item }">
      <AppBadge :variant="severityVariant(item.entry.binding?.severity)" dot size="sm">
        {{ severityLabel(item.entry.binding?.severity) }}
      </AppBadge>
    </template>

    <template #cell(effective)="{ item }">
      <span class="mono">{{ effectiveWindowLabel(item.entry) }}</span>
    </template>

    <template #cell(status)="{ item }">
      <AppBadge :variant="stateVariant(item.entry)" dot size="sm">
        {{ stateLabel(item.entry) }}
      </AppBadge>
    </template>

    <template #cell(actions)="{ item }">
      <AppButton
        v-if="canManage"
        variant="danger"
        size="sm"
        :disabled="busy[item.entry.rule_instance_id ?? '']"
        @click="emit('deactivate', item.entry)"
      >
        {{ t("portfolio.compliance.deactivate") }}
      </AppButton>
      <span v-else class="rule-cell__readonly">{{ t("portfolio.compliance.readOnly") }}</span>
    </template>
  </AppDataTable>
  </AppSection>
</template>

<style scoped>
.rule-cell__label {
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.rule-cell__explanation {
  margin-top: 2px;
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  max-width: 28rem;
}

.rule-cell__type-id {
  display: inline-block;
  margin-top: 4px;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-2xs);
  color: var(--text-tertiary);
}

.rule-cell__readonly {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  font-style: italic;
}

.param-list {
  margin: 0;
  display: grid;
  gap: 2px;
  font-size: var(--font-size-xs);
}

.param-list__row {
  display: flex;
  gap: var(--space-2);
}

.param-list__row dt {
  color: var(--text-tertiary);
}

.param-list__row dd {
  margin: 0;
  color: var(--text-primary);
  font-weight: var(--font-weight-medium);
}

.mono {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
}
</style>
