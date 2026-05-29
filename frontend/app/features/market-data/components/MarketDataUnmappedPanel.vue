<script setup lang="ts">
import type {
  ApiSecurity,
  ApiUnmappedCandidate,
} from "../services/marketDataCatalogApi";
import { useI18n } from "~/composables/useI18n";

const props = defineProps<{
  items: ApiUnmappedCandidate[];
  loading?: boolean;
  securities: ApiSecurity[];
}>();

const emit = defineEmits<{
  (e: "map", candidateId: string, securityId: string): void;
  (e: "reject", candidateId: string, reason: string): void;
}>();
const { t } = useI18n();

// Per-row UI state — kept in maps keyed by candidate_id so we don't lose
// selections across re-renders.
const selectedSecurity = reactive<Record<string, string>>({});
const rejectReason = reactive<Record<string, string>>({});

function suggestedFor(c: ApiUnmappedCandidate): string {
  return c.suggested_security_id ?? "";
}

function onMap(c: ApiUnmappedCandidate) {
  const id = c.candidate_id;
  if (!id) return;
  const target = selectedSecurity[id] || suggestedFor(c);
  if (!target) return;
  emit("map", id, target);
}

function onReject(c: ApiUnmappedCandidate) {
  const id = c.candidate_id;
  if (!id) return;
  const reason = rejectReason[id]?.trim();
  if (!reason) return;
  emit("reject", id, reason);
}

const securityOptions = computed(() =>
  props.securities
    .filter((s) => s.security_id)
    .map((s) => ({
      value: s.security_id ?? "",
      label: `${s.ims_symbol ?? s.security_id} — ${s.name ?? ""}`,
    })),
);
</script>

<template>
  <div class="card mdc-candidates">
    <div class="card-header mdc-candidates__head">
      <span class="card-title">{{ t("marketData.headings.unmappedProviderSymbols") }}</span>
      <span class="mdc-count">{{ t("marketData.messages.pendingReview", { count: items.length }) }}</span>
    </div>

    <div v-if="items.length === 0 && !loading" class="mdc-empty">
      {{ t("marketData.messages.importFullyResolved") }}
    </div>

    <div v-else class="mdc-candidate-list">
      <div v-for="c in items" :key="c.candidate_id" class="mdc-candidate">
        <div class="mdc-candidate__head">
          <div>
            <span class="mdc-candidate__provider">{{ c.provider_code }}</span>
            <span class="mdc-candidate__symbol">{{ c.provider_symbol }}</span>
            <span v-if="c.provider_asset_type" class="mdc-candidate__meta">· {{ c.provider_asset_type }}</span>
            <span v-if="c.provider_currency" class="mdc-candidate__meta">· {{ c.provider_currency }}</span>
            <span v-if="c.isin" class="mdc-candidate__meta">· ISIN {{ c.isin }}</span>
          </div>
          <span class="badge badge-warning">{{ c.candidate_status }}</span>
        </div>

        <div v-if="c.provider_name" class="mdc-candidate__line">
          <span class="mdc-muted">{{ t("marketData.labels.providerName") }}</span> {{ c.provider_name }}
        </div>

        <div class="mdc-candidate__actions">
          <label class="mdc-field">
            <span class="mdc-label">{{ t("marketData.labels.mapToSecurity") }}</span>
            <select
              :value="selectedSecurity[c.candidate_id ?? ''] || suggestedFor(c)"
              class="form-input mdc-candidate__select"
              @change="selectedSecurity[c.candidate_id ?? ''] = ($event.target as HTMLSelectElement).value"
            >
              <option value="">{{ t("marketData.placeholders.pickSecurity") }}</option>
              <option
                v-for="opt in securityOptions"
                :key="opt.value"
                :value="opt.value"
              >{{ opt.label }}</option>
            </select>
          </label>
          <button
            class="btn btn-primary btn-sm mdc-candidate__btn"
            type="button"
            :disabled="!(selectedSecurity[c.candidate_id ?? ''] || suggestedFor(c))"
            @click="onMap(c)"
          >{{ t("marketData.actions.map") }}</button>
        </div>

        <div class="mdc-candidate__actions">
          <label class="mdc-field">
            <span class="mdc-label">{{ t("marketData.labels.rejectReason") }}</span>
            <input
              :value="rejectReason[c.candidate_id ?? ''] ?? ''"
              class="form-input"
              :placeholder="t('marketData.placeholders.rejectReason')"
              @input="rejectReason[c.candidate_id ?? ''] = ($event.target as HTMLInputElement).value"
            />
          </label>
          <button
            class="btn btn-secondary btn-sm mdc-candidate__btn"
            type="button"
            :disabled="!(rejectReason[c.candidate_id ?? '']?.trim())"
            @click="onReject(c)"
          >{{ t("marketData.actions.reject") }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.mdc-candidates__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.mdc-count {
  font-size: 0.75rem;
  color: var(--md-text-muted);
}
.mdc-candidate-list {
  display: grid;
  gap: 0.75rem;
  padding: 0.75rem;
}
.mdc-candidate {
  border: 1px solid var(--md-border);
  border-radius: 0.5rem;
  padding: 0.75rem;
  background: var(--md-surface);
  display: grid;
  gap: 0.5rem;
}
.mdc-candidate__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}
.mdc-candidate__provider {
  display: inline-block;
  padding: 0.05rem 0.4rem;
  border-radius: 0.25rem;
  background: var(--md-surface-muted);
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-right: 0.4rem;
}
.mdc-candidate__symbol {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-weight: 600;
}
.mdc-candidate__meta {
  margin-left: 0.4rem;
  font-size: 0.78rem;
  color: var(--md-text-muted);
}
.mdc-candidate__line {
  font-size: 0.85rem;
}
.mdc-muted {
  color: var(--md-text-muted);
}
.mdc-candidate__actions {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 0.5rem;
  align-items: end;
}
.mdc-candidate__select,
.mdc-candidate__btn {
  min-width: 0;
}
.mdc-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.mdc-label {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--md-text-muted);
}
.mdc-empty {
  padding: 1.5rem;
  text-align: center;
  color: var(--md-text-muted);
}
</style>
