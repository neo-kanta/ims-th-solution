<script setup lang="ts">
import type { ProviderMappingView } from "../market-data.types";
import { useI18n } from "~/composables/useI18n";

defineProps<{
  mappings: ProviderMappingView[];
  /** Kept for future deep-linking; not used as a URL anymore. */
  securityId?: string;
}>();

const emit = defineEmits<{
  (e: "manage"): void;
  (e: "open-unmapped"): void;
}>();
const { t } = useI18n();

function statusVariant(status?: string): string {
  switch (status) {
    case "ACTIVE":   return "md-status-badge--fresh";
    case "INACTIVE": return "md-sec-mappings__pill--neutral";
    default:         return "md-status-badge--stale";
  }
}
</script>

<template>
  <article class="card md-sec-card">
    <header class="card-header md-sec-mappings__head">
      <span class="card-title">{{ t("marketData.labels.providerMappings") }}</span>
      <button
        type="button"
        class="md-sec-mappings__link"
        @click="emit('manage')"
      >{{ t("marketData.actions.manageMappings") }}</button>
    </header>

    <div class="card-body">
      <div v-if="mappings.length === 0" class="md-sec-mappings__empty">
        <p class="md-sec-mappings__empty-title">{{ t("marketData.messages.noProviderMapping") }}</p>
        <p class="md-sec-mappings__empty-text">
          {{ t("marketData.messages.syncActionsNeedMapping") }}
        </p>
        <div class="md-sec-mappings__empty-actions">
          <button
            type="button"
            class="btn btn-primary btn-sm"
            @click="emit('manage')"
          >{{ t("marketData.actions.addProviderMapping") }}</button>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            @click="emit('open-unmapped')"
          >{{ t("marketData.actions.openUnmappedQueue") }}</button>
        </div>
      </div>

      <div v-else class="md-sec-mappings__table-wrap">
        <table class="data-table md-sec-mappings__table">
          <thead>
            <tr>
              <th>{{ t("marketData.labels.provider") }}</th>
              <th>{{ t("marketData.labels.providerSymbol") }}</th>
              <th>{{ t("marketData.labels.exchange") }}</th>
              <th>{{ t("marketData.labels.ccy") }}</th>
              <th class="col-right">{{ t("marketData.labels.priority") }}</th>
              <th>{{ t("marketData.labels.status") }}</th>
              <th>{{ t("marketData.labels.primary") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="m in mappings" :key="m.mappingId">
              <td>
                <span class="md-provider-pill md-provider-pill--xs" :data-provider="m.providerCode">
                  {{ m.providerCode === "alpha_vantage" ? "AV" : m.providerCode === "yahoo" ? "YF" : m.providerCode.slice(0, 2).toUpperCase() }}
                </span>
                <span class="md-sec-mappings__provider-label">{{ m.providerCode }}</span>
              </td>
              <td><code>{{ m.providerSymbol }}</code></td>
              <td>{{ m.providerExchange || "—" }}</td>
              <td>{{ m.providerCurrency || "—" }}</td>
              <td class="col-right md-num">{{ m.priority ?? "—" }}</td>
              <td>
                <span
                  class="md-status-badge"
                  :class="statusVariant(m.status)"
                  :aria-label="`Mapping status ${m.status}`"
                >
                  <span class="md-status-dot" aria-hidden="true" />
                  {{ m.status }}
                </span>
              </td>
              <td>
                <span v-if="m.isPrimary" class="md-status-badge md-status-badge--fresh">
                  <span class="md-status-dot" aria-hidden="true" />
                  Yes
                </span>
                <span v-else class="md-text-muted">{{ t("marketData.labels.no") }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </article>
</template>

<style scoped>
.md-sec-mappings__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.md-sec-mappings__link {
  font-size: 0.85rem;
  color: var(--md-accent, var(--md-text-muted));
  background: transparent;
  border: 0;
  padding: 0;
  cursor: pointer;
  font: inherit;
  font-size: 0.85rem;
}
.md-sec-mappings__link:hover { text-decoration: underline; }
.md-sec-mappings__link:focus-visible {
  outline: 2px solid var(--md-accent);
  outline-offset: 2px;
  border-radius: 2px;
}
.md-sec-mappings__empty {
  display: grid;
  gap: 0.4rem;
  padding: 0.5rem 0;
}
.md-sec-mappings__empty-title {
  margin: 0;
  font-weight: 600;
}
.md-sec-mappings__empty-text {
  margin: 0;
  color: var(--md-text-muted);
  font-size: 0.85rem;
}
.md-sec-mappings__empty-actions {
  display: flex;
  gap: 0.4rem;
  flex-wrap: wrap;
  margin-top: 0.4rem;
}
.md-sec-mappings__table-wrap {
  overflow-x: auto;
}
.md-sec-mappings__table {
  width: 100%;
  font-size: 0.85rem;
}
.md-sec-mappings__provider-label {
  margin-left: 0.4rem;
  font-size: 0.82rem;
  color: var(--md-text-muted);
}
.md-sec-mappings__pill--neutral {
  background: var(--md-neutral-bg);
  color: var(--md-neutral-text);
}
.md-num { font-variant-numeric: tabular-nums; }
.col-right { text-align: right; }
</style>
