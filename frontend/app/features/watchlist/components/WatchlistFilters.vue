<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "~/composables/useI18n";
import type { ItemListQuery, AlertListQuery } from "../types";
import type { WatchlistPortfolio } from "../composables/useWatchlistPortfolios";
import type { WatchlistSecurity } from "../composables/useWatchlistSecuritySearch";

export interface WatchlistFilterState {
  items: ItemListQuery;
  alerts: AlertListQuery;
}

interface Props {
  modelValue: WatchlistFilterState;
  portfolios: WatchlistPortfolio[];
  portfoliosLoading?: boolean;
  portfoliosError?: string | null;
  securitySearchResults?: WatchlistSecurity[];
  securitySearchLoading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  portfoliosLoading: false,
  portfoliosError: null,
  securitySearchResults: () => [],
  securitySearchLoading: false,
});

const emit = defineEmits<{
  "update:modelValue": [value: WatchlistFilterState];
  "security-search": [query: string];
}>();

const { t } = useI18n();

const securityQuery = ref("");
const showSecurityDropdown = ref(false);

function updateItems(partial: Partial<ItemListQuery>) {
  emit("update:modelValue", {
    items: { ...props.modelValue.items, ...partial },
    alerts: { ...props.modelValue.alerts },
  });
}

function updateAlerts(partial: Partial<AlertListQuery>) {
  emit("update:modelValue", {
    items: { ...props.modelValue.items },
    alerts: { ...props.modelValue.alerts, ...partial },
  });
}

function onAckFilter(e: Event) {
  const val = (e.target as HTMLSelectElement).value;
  if (val === "") {
    updateAlerts({ acknowledged: undefined });
  } else {
    updateAlerts({ acknowledged: val === "true" });
  }
}

function onScopeFilterChange(val: string) {
  const scope = val ? (val as "PERSONAL" | "PORTFOLIO") : undefined;
  emit("update:modelValue", {
    items: { ...props.modelValue.items, scope_type: scope, portfolio_id: undefined },
    alerts: { ...props.modelValue.alerts, scope_type: scope, portfolio_id: undefined },
  });
}

function onPortfolioFilterChange(val: string) {
  const pId = val || undefined;
  emit("update:modelValue", {
    items: { ...props.modelValue.items, portfolio_id: pId },
    alerts: { ...props.modelValue.alerts, portfolio_id: pId },
  });
}

function onSecuritySearch(e: Event) {
  const q = (e.target as HTMLInputElement).value;
  securityQuery.value = q;
  showSecurityDropdown.value = true;
  emit("security-search", q);
}

function selectSecurity(sec: WatchlistSecurity) {
  securityQuery.value = sec.display_symbol ?? sec.ims_symbol ?? "";
  showSecurityDropdown.value = false;
  emit("update:modelValue", {
    items: { ...props.modelValue.items, security_id: sec.security_id },
    alerts: { ...props.modelValue.alerts, security_id: sec.security_id },
  });
}

function clearSecurityFilter() {
  securityQuery.value = "";
  emit("update:modelValue", {
    items: { ...props.modelValue.items, security_id: undefined },
    alerts: { ...props.modelValue.alerts, security_id: undefined },
  });
}

function onBlur() {
  setTimeout(() => {
    showSecurityDropdown.value = false;
  }, 200);
}
</script>

<template>
  <div class="watchlist-filters">
    <div class="form-grid-3">
      <!-- Item Status -->
      <div class="form-group">
        <label class="form-label">{{ t('watchlist.item.status') }}</label>
        <select
          class="form-select"
          :value="modelValue.items.include_disabled ? 'all' : 'active'"
          @change="updateItems({ include_disabled: ($event.target as HTMLSelectElement).value === 'all' })"
        >
          <option value="active">{{ t('watchlist.status.active') }}</option>
          <option value="all">{{ t('watchlist.filters.includeDisabled') }}</option>
        </select>
      </div>

      <!-- Scope Filter -->
      <div class="form-group">
        <label class="form-label">{{ t('watchlist.filters.scope') }}</label>
        <select
          class="form-select"
          :value="modelValue.items.scope_type ?? ''"
          @change="onScopeFilterChange(($event.target as HTMLSelectElement).value)"
        >
          <option value="">{{ t('watchlist.scope.all') }}</option>
          <option value="PERSONAL">{{ t('watchlist.scope.personal') }}</option>
          <option value="PORTFOLIO">{{ t('watchlist.scope.portfolio') }}</option>
        </select>
      </div>

      <!-- Portfolio Filter (visible only when scope is PORTFOLIO) -->
      <div v-if="modelValue.items.scope_type === 'PORTFOLIO'" class="form-group">
        <label class="form-label">{{ t('watchlist.filters.portfolio') }}</label>
        <div v-if="portfoliosError" class="form-helper" style="color: var(--color-error, red);">
          {{ portfoliosError }}
        </div>
        <select
          v-else
          class="form-select"
          :value="modelValue.items.portfolio_id ?? ''"
          :disabled="portfoliosLoading"
          @change="onPortfolioFilterChange(($event.target as HTMLSelectElement).value)"
        >
          <option value="">{{ t('watchlist.filters.portfolioPlaceholder') }}</option>
          <option
            v-for="p in portfolios"
            :key="p.id"
            :value="p.id"
          >
            {{ p.code ?? p.id }} – {{ p.name ?? '' }}
          </option>
        </select>
      </div>

      <!-- Security Filter -->
      <div class="form-group" style="position: relative;">
        <label class="form-label">{{ t('watchlist.filters.security') }}</label>
        <div style="display: flex; gap: 8px; align-items: center;">
          <input
            class="form-input"
            type="text"
            :placeholder="t('watchlist.drawer.securitySearch')"
            :value="securityQuery"
            :disabled="securitySearchLoading"
            @input="onSecuritySearch"
            @focus="showSecurityDropdown = true"
            @blur="onBlur"
          />
          <button
            v-if="modelValue.items.security_id"
            class="btn btn-secondary btn-sm"
            type="button"
            @click="clearSecurityFilter"
          >
            {{ t('watchlist.filters.clear') }}
          </button>
        </div>
        <div
          v-if="showSecurityDropdown && securitySearchResults.length > 0"
          style="position: absolute; top: 100%; left: 0; right: 0; z-index: 100; background: white; border: 1px solid var(--color-border, #e2e8f0); border-radius: 4px; max-height: 200px; overflow-y: auto; box-shadow: 0 4px 12px rgba(0,0,0,0.1);"
        >
          <button
            v-for="sec in securitySearchResults"
            :key="sec.security_id"
            type="button"
            style="display: block; width: 100%; text-align: left; padding: 8px 12px; border: none; background: none; cursor: pointer; font-size: 13px;"
            @mousedown.prevent="selectSecurity(sec)"
          >
            <span style="font-weight: 600;">{{ sec.display_symbol ?? sec.ims_symbol }}</span>
            <span style="color: var(--color-text-muted, #64748b); margin-left: 8px;">{{ sec.name }}</span>
          </button>
        </div>
      </div>

      <!-- Alert Acknowledgement -->
      <div class="form-group">
        <label class="form-label">{{ t('watchlist.filters.acknowledged') }}</label>
        <select
          class="form-select"
          :value="modelValue.alerts.acknowledged === undefined ? '' : String(modelValue.alerts.acknowledged)"
          @change="onAckFilter"
        >
          <option value="">{{ t('watchlist.filters.allAlerts') }}</option>
          <option value="false">{{ t('watchlist.filters.onlyUnacked') }}</option>
          <option value="true">{{ t('watchlist.filters.onlyAcked') }}</option>
        </select>
      </div>

      <!-- Alerts From -->
      <div class="form-group">
        <label class="form-label">{{ t('watchlist.filters.dateFrom') }}</label>
        <input
          class="form-input"
          type="date"
          :value="modelValue.alerts.created_from ? modelValue.alerts.created_from.slice(0, 10) : ''"
          @change="updateAlerts({ created_from: ($event.target as HTMLInputElement).value ? ($event.target as HTMLInputElement).value + 'T00:00:00Z' : undefined })"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.form-grid-3 {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
}
@media (max-width: 768px) {
  .form-grid-3 {
    grid-template-columns: 1fr;
  }
}
</style>
