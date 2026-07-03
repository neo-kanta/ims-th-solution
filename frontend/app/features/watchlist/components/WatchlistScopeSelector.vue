<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import type { WatchlistPortfolio } from "../composables/useWatchlistPortfolios";
import { portfolioLabel } from "../lib/formatters";

interface Props {
  scopeType: "PERSONAL" | "PORTFOLIO";
  portfolioId?: string | null;
  portfolios: WatchlistPortfolio[];
  portfoliosLoading?: boolean;
  portfoliosError?: string | null;
  disabled?: boolean;
}

withDefaults(defineProps<Props>(), {
  portfolioId: null,
  portfoliosLoading: false,
  portfoliosError: null,
  disabled: false,
});

const emit = defineEmits<{
  "update:scopeType": [value: "PERSONAL" | "PORTFOLIO"];
  "update:portfolioId": [value: string | null];
}>();

const { t } = useI18n();

function onScopeChange(e: Event) {
  const val = (e.target as HTMLSelectElement).value as "PERSONAL" | "PORTFOLIO";
  emit("update:scopeType", val);
  emit("update:portfolioId", null);
}
</script>

<template>
  <div class="watchlist-scope-selector">
    <div class="form-grid-2">
      <div class="form-group">
        <label class="form-label">{{ t('watchlist.item.scope') }}</label>
        <select
          class="form-select"
          :value="scopeType"
          :disabled="disabled"
          @change="onScopeChange"
        >
          <option value="PERSONAL">{{ t('watchlist.scope.personal') }}</option>
          <option value="PORTFOLIO">{{ t('watchlist.scope.portfolio') }}</option>
        </select>
      </div>
      <div v-if="scopeType === 'PORTFOLIO'" class="form-group">
        <label class="form-label form-label-req">{{ t('watchlist.filters.portfolio') }}</label>
        <div v-if="portfoliosError" class="form-helper" style="color: var(--color-error, red);">
          {{ portfoliosError }}
        </div>
        <select
          v-else
          class="form-select"
          :value="portfolioId ?? ''"
          :disabled="disabled || portfoliosLoading"
          @change="emit('update:portfolioId', ($event.target as HTMLSelectElement).value || null)"
        >
          <option value="">{{ t('watchlist.filters.portfolioPlaceholder') }}</option>
          <option
            v-for="p in portfolios"
            :key="p.id"
            :value="p.id"
          >
            {{ portfolioLabel(p) }}
          </option>
        </select>
      </div>
    </div>
  </div>
</template>

<style scoped>
.form-grid-2 {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}
@media (max-width: 640px) {
  .form-grid-2 {
    grid-template-columns: 1fr;
  }
}
</style>
