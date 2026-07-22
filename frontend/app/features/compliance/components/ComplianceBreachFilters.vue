<script setup lang="ts">
/**
 * Compact filter toolbar for the Breaches queue — replaces the former
 * permanent left filter rail. Every change applies immediately (no separate
 * Apply step): a status click, a portfolio/rule pick, or a date edit updates
 * `modelValue` right away, and the page shell resets pagination and refetches
 * in response.
 *
 * Contract filtering is intentionally omitted here — the only contract
 * identity available client-side today is a raw UUID list
 * (`authStore.permissions.contracts`) with no human-readable label, and the
 * UX requirement is that a filter must resolve to a business label before it
 * can be exposed. See the Breaches page report for this backend gap.
 */
import { computed, onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppDateRangeField from "~/shared/ui/AppDateRangeField.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

import { useCompliancePortfolioDirectory } from "../composables/useCompliancePortfolioDirectory";
import {
  defaultBreachFilterState,
  hasNonDefaultFilters,
  type BreachFilterState,
  type BreachStatusChoice,
} from "../lib/breachFilters";
import { RULE_CATALOG } from "../lib/ruleTypeCatalog";
import ComplianceFilterCombobox, {
  type ComplianceComboboxOption,
} from "./ComplianceFilterCombobox.vue";

const state = defineModel<BreachFilterState>({ required: true });

defineProps<{ loading?: boolean }>();

const { t } = useI18n();
const portfolios = useCompliancePortfolioDirectory();

onMounted(() => {
  void portfolios.ensureLoaded();
});

const statusOptions = computed<{ value: BreachStatusChoice; label: string }[]>(() => [
  { value: "OPEN", label: t("compliance.postTrade.toolbar.statusOpen") },
  { value: "OVERRIDDEN", label: t("compliance.postTrade.toolbar.statusOverridden") },
  { value: "RESOLVED", label: t("compliance.postTrade.toolbar.statusResolved") },
  { value: "ALL", label: t("compliance.postTrade.toolbar.statusAll") },
]);

function setStatus(next: BreachStatusChoice) {
  if (state.value.status === next) return;
  state.value = { ...state.value, status: next };
}

const portfolioOptions = computed<ComplianceComboboxOption[]>(() =>
  portfolios.items.value.map((p) => ({
    value: p.id,
    label: p.code,
    sublabel: p.name,
  })),
);

function setPortfolio(portfolioId: string) {
  state.value = { ...state.value, portfolioId };
}

const ruleOptions = computed(() => {
  const options = RULE_CATALOG.map((entry) => ({
    value: entry.typeId,
    label: t(entry.labelKey),
  })).sort((a, b) => a.label.localeCompare(b.label));

  const current = state.value.ruleTypeId;
  if (current && !options.some((o) => o.value === current)) {
    options.unshift({ value: current, label: current });
  }
  return options;
});

function onRuleChange(event: Event) {
  state.value = { ...state.value, ruleTypeId: (event.target as HTMLSelectElement).value };
}

const dateFrom = computed({
  get: () => state.value.dateFrom,
  set: (value: string) => {
    state.value = { ...state.value, dateFrom: value };
  },
});
const dateTo = computed({
  get: () => state.value.dateTo,
  set: (value: string) => {
    state.value = { ...state.value, dateTo: value };
  },
});

interface Chip {
  key: string;
  label: string;
  remove: () => void;
}

const chips = computed<Chip[]>(() => {
  const list: Chip[] = [];
  const s = state.value;

  if (s.ruleTypeId) {
    const entry = RULE_CATALOG.find((r) => r.typeId === s.ruleTypeId);
    const label = t("compliance.postTrade.toolbar.chipRule", {
      label: entry ? t(entry.labelKey) : s.ruleTypeId,
    });
    list.push({ key: "rule", label, remove: () => setRuleTypeId("") });
  }

  if (s.portfolioId) {
    const portfolio = portfolios.byId.value.get(s.portfolioId);
    const label = t("compliance.postTrade.toolbar.chipPortfolio", {
      label: portfolio ? portfolio.code : s.portfolioId,
    });
    list.push({ key: "portfolio", label, remove: () => setPortfolio("") });
  }

  if (s.dateFrom && s.dateTo) {
    list.push({
      key: "dateRange",
      label: t("compliance.postTrade.toolbar.chipDateRange", { from: s.dateFrom, to: s.dateTo }),
      remove: () => {
        state.value = { ...state.value, dateFrom: "", dateTo: "" };
      },
    });
  } else if (s.dateFrom) {
    list.push({
      key: "dateFrom",
      label: t("compliance.postTrade.toolbar.chipDateFrom", { date: s.dateFrom }),
      remove: () => {
        state.value = { ...state.value, dateFrom: "" };
      },
    });
  } else if (s.dateTo) {
    list.push({
      key: "dateTo",
      label: t("compliance.postTrade.toolbar.chipDateTo", { date: s.dateTo }),
      remove: () => {
        state.value = { ...state.value, dateTo: "" };
      },
    });
  }

  return list;
});

function setRuleTypeId(ruleTypeId: string) {
  state.value = { ...state.value, ruleTypeId };
}

const showClearAll = computed(() => hasNonDefaultFilters(state.value));

function clearAll() {
  state.value = defaultBreachFilterState();
}
</script>

<template>
  <div class="breach-toolbar">
    <div class="breach-toolbar__row">
      <div
        class="breach-toolbar__segment"
        role="radiogroup"
        :aria-label="t('compliance.postTrade.toolbar.statusLabel')"
      >
        <button
          v-for="opt in statusOptions"
          :key="opt.value"
          type="button"
          role="radio"
          :aria-checked="state.status === opt.value"
          class="breach-toolbar__segment-btn"
          :class="{ 'is-active': state.status === opt.value }"
          :disabled="loading"
          @click="setStatus(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>

      <div class="breach-toolbar__field breach-toolbar__field--portfolio">
        <label class="breach-toolbar__label" for="breach-toolbar-portfolio">
          {{ t("compliance.postTrade.toolbar.portfolioLabel") }}
        </label>
        <ComplianceFilterCombobox
          input-id="breach-toolbar-portfolio"
          :model-value="state.portfolioId"
          :options="portfolioOptions"
          :disabled="!!portfolios.error.value"
          :loading="portfolios.loading.value"
          :placeholder="t('compliance.postTrade.toolbar.portfolioPlaceholder')"
          :field-label="t('compliance.postTrade.toolbar.portfolioLabel')"
          :loading-text="t('compliance.postTrade.toolbar.portfolioLoading')"
          :empty-text="t('compliance.postTrade.toolbar.portfolioEmpty')"
          :change-label="t('compliance.postTrade.toolbar.portfolioChange')"
          @update:model-value="setPortfolio"
        />
      </div>

      <div class="breach-toolbar__field">
        <label class="breach-toolbar__label" for="breach-toolbar-rule">
          {{ t("compliance.postTrade.toolbar.ruleLabel") }}
        </label>
        <select
          id="breach-toolbar-rule"
          class="form-control"
          :value="state.ruleTypeId"
          @change="onRuleChange"
        >
          <option value="">{{ t("compliance.postTrade.toolbar.ruleAny") }}</option>
          <option v-for="opt in ruleOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div class="breach-toolbar__field">
        <span class="breach-toolbar__label">
          {{ t("compliance.postTrade.toolbar.dateRangeLabel") }}
        </span>
        <AppDateRangeField v-model:start="dateFrom" v-model:end="dateTo" />
      </div>
    </div>

    <p v-if="portfolios.error.value" class="breach-toolbar__notice" role="status">
      {{ t("compliance.postTrade.toolbar.portfolioUnavailable") }}
    </p>

    <div v-if="chips.length || showClearAll" class="breach-toolbar__chips">
      <span class="breach-toolbar__chips-label">
        {{ t("compliance.postTrade.toolbar.activeFilters") }}
      </span>
      <button
        v-for="chip in chips"
        :key="chip.key"
        type="button"
        class="breach-toolbar__chip"
        :aria-label="`${t('compliance.postTrade.toolbar.removeFilter')}: ${chip.label}`"
        @click="chip.remove"
      >
        {{ chip.label }}
        <AppIcon name="close" size="xs" />
      </button>
      <button
        v-if="showClearAll"
        type="button"
        class="breach-toolbar__clear"
        @click="clearAll"
      >
        {{ t("compliance.postTrade.toolbar.clearAll") }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.breach-toolbar {
  display: grid;
  gap: var(--space-3);
}

.breach-toolbar__row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: var(--space-4);
}

.breach-toolbar__segment {
  display: inline-flex;
  gap: 2px;
  padding: 3px;
  background: var(--bg-card-muted);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  align-self: flex-end;
}

.breach-toolbar__segment-btn {
  border: none;
  background: transparent;
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--text-secondary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  white-space: nowrap;
}

.breach-toolbar__segment-btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.breach-toolbar__segment-btn.is-active {
  background: var(--action-primary, #0969da);
  color: #fff;
}

.breach-toolbar__field {
  display: grid;
  gap: var(--space-1);
  min-width: 12rem;
}

.breach-toolbar__field--portfolio {
  min-width: 16rem;
}

.breach-toolbar__label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.form-control {
  width: 100%;
  height: var(--size-control-md, 36px);
  padding: 0 var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.form-control:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.breach-toolbar__notice {
  margin: 0;
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--alert-warning-border, #d4a72c);
  border-radius: var(--radius-md);
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  font-size: var(--font-size-xs);
}

.breach-toolbar__chips {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
}

.breach-toolbar__chips-label {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.breach-toolbar__chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  height: 26px;
  padding: 0 var(--space-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-pill);
  background: var(--bg-card-muted);
  color: var(--text-primary);
  font-size: var(--font-size-xs);
  cursor: pointer;
}

.breach-toolbar__chip:hover {
  background: var(--bg-card-hover, #f0f2f4);
}

.breach-toolbar__clear {
  border: none;
  background: transparent;
  color: var(--action-primary, #0969da);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  cursor: pointer;
  padding: 0 var(--space-1);
}

@media (max-width: 768px) {
  .breach-toolbar__row {
    flex-direction: column;
    align-items: stretch;
  }

  .breach-toolbar__segment {
    align-self: stretch;
    justify-content: space-between;
  }

  .breach-toolbar__segment-btn {
    flex: 1;
  }

  .breach-toolbar__field,
  .breach-toolbar__field--portfolio {
    min-width: 0;
  }
}
</style>
