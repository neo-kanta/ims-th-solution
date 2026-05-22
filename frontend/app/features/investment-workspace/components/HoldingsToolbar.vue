<script setup lang="ts">
import AppButton from "~/shared/ui/AppButton.vue";
import { useI18n } from "~/composables/useI18n";
import type { FundCard, Freshness } from "../types";

defineProps<{
  funds: FundCard[];
  activeFundId: string;
  asOf: string;
  loading: boolean;
  exporting: boolean;
  freshness: Freshness | null;
}>();

const emit = defineEmits<{
  changeFund: [fundId: string];
  changeAsOf: [date: string];
  refresh: [];
  export: [];
}>();

const { t } = useI18n();

function onChangeFund(event: Event) {
  const target = event.target as HTMLSelectElement | null;
  if (target) emit("changeFund", target.value);
}
function onChangeAsOf(event: Event) {
  const target = event.target as HTMLInputElement | null;
  if (target?.value) emit("changeAsOf", target.value);
}
</script>

<template>
  <div class="ht-toolbar">
    <div class="ht-toolbar__field">
      <label class="ht-toolbar__label" for="ht-contract">
        {{ t("holdings.toolbar.contract", "CONTRACT") }}
      </label>
      <select
        id="ht-contract"
        class="ht-toolbar__select"
        :value="activeFundId"
        @change="onChangeFund"
      >
        <option v-for="f in funds" :key="f.fund_id" :value="f.fund_id">
          {{ f.contract_code }} · {{ f.short_name }}
        </option>
      </select>
    </div>

    <div class="ht-toolbar__field">
      <label class="ht-toolbar__label" for="ht-asof">
        {{ t("holdings.toolbar.asOf", "AS OF") }}
      </label>
      <input
        id="ht-asof"
        class="ht-toolbar__input"
        type="date"
        :value="asOf"
        @change="onChangeAsOf"
      />
    </div>

    <div class="ht-toolbar__actions">
      <AppButton variant="secondary" size="sm" :loading="loading" @click="emit('refresh')">
        {{ loading ? t("holdings.toolbar.refreshing", "Refreshing") : t("holdings.toolbar.refresh", "Refresh") }}
      </AppButton>
      <AppButton variant="primary" size="sm" :loading="exporting" @click="emit('export')">
        {{ exporting ? t("holdings.toolbar.exporting", "Exporting") : t("holdings.toolbar.export", "Export") }}
      </AppButton>

      <span
        v-if="freshness"
        class="ht-toolbar__freshness"
        :class="freshness.is_stale ? 'is-stale' : 'is-fresh'"
        :title="freshness.label"
      >
        <span class="ht-toolbar__freshness-dot" aria-hidden="true" />
        {{
          freshness.is_stale
            ? t("holdings.toolbar.freshStale", "Stale data")
            : t("holdings.toolbar.freshOk", "Feeds OK")
        }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.ht-toolbar {
  display: flex;
  align-items: end;
  gap: var(--space-3);
  padding: var(--space-3);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  flex-wrap: wrap;
}

.ht-toolbar__field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.ht-toolbar__label {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}

.ht-toolbar__select,
.ht-toolbar__input {
  height: 30px;
  padding: 0 10px;
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-primary);
  font-family: inherit;
  min-width: 180px;
}

.ht-toolbar__input {
  min-width: 160px;
}

.ht-toolbar__select:focus,
.ht-toolbar__input:focus {
  outline: 2px solid var(--focus-ring);
  outline-offset: -1px;
  border-color: var(--border-focus);
}

.ht-toolbar__actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-left: auto;
}

.ht-toolbar__freshness {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 6px;
  border: 1px solid transparent;
}

.ht-toolbar__freshness.is-fresh {
  background: var(--status-approved-bg);
  border-color: var(--border-success);
  color: var(--status-approved-text);
}

.ht-toolbar__freshness.is-stale {
  background: var(--status-pending-bg);
  border-color: var(--border-warning);
  color: var(--status-pending-text);
}

.ht-toolbar__freshness-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
}

@media (max-width: 640px) {
  .ht-toolbar__actions {
    margin-left: 0;
    width: 100%;
    justify-content: flex-start;
    flex-wrap: wrap;
  }
}
</style>
